package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/pion/webrtc/v3"
)

const (
	msgChSize = 20
)

type session struct {
	mut sync.RWMutex

	userID    string
	channelID string
	connID    string

	// WebSocket
	signalInCh  chan []byte
	signalOutCh chan []byte
	wsMsgCh     chan clientMessage
	wsCloseCh   chan struct{}
	doneCh      chan struct{}

	// WebRTC
	outVoiceTrack        *webrtc.TrackLocalStaticRTP
	outVoiceTrackEnabled bool
	outScreenTrack       *webrtc.TrackLocalStaticRTP
	remoteScreenTrack    *webrtc.TrackRemote
	rtcConn              *webrtc.PeerConnection
	tracksCh             chan *webrtc.TrackLocalStaticRTP
	iceCh                chan []byte
	closeCh              chan struct{}

	trackEnableCh chan bool
	rtpSendersMap map[*webrtc.TrackLocalStaticRTP]*webrtc.RTPSender
}

func newUserSession(userID, channelID, connID string) *session {
	return &session{
		userID:        userID,
		channelID:     channelID,
		connID:        connID,
		signalInCh:    make(chan []byte, msgChSize),
		signalOutCh:   make(chan []byte, msgChSize),
		wsMsgCh:       make(chan clientMessage, msgChSize),
		wsCloseCh:     make(chan struct{}),
		tracksCh:      make(chan *webrtc.TrackLocalStaticRTP, 5),
		iceCh:         make(chan []byte, msgChSize),
		closeCh:       make(chan struct{}),
		doneCh:        make(chan struct{}),
		trackEnableCh: make(chan bool, 5),
		rtpSendersMap: map[*webrtc.TrackLocalStaticRTP]*webrtc.RTPSender{},
	}
}

func (p *Plugin) addUserSession(userID, channelID string, userSession *session) (channelState, error) {
	var st channelState
	err := p.kvSetAtomicChannelState(channelID, func(state *channelState) (*channelState, error) {
		if state == nil {
			return nil, fmt.Errorf("channel state is missing from store")
		}
		if state.Call == nil {
			state.Call = &callState{
				ID:      model.NewId(),
				StartAt: time.Now().UnixMilli(),
				Users:   make(map[string]*userState),
			}
			state.NodeID = p.nodeID
		}

		if _, ok := state.Call.Users[userID]; ok {
			return nil, fmt.Errorf("user is already connected")
		}
		state.Call.Users[userID] = &userState{}

		st = *state
		return state, nil
	})

	return st, err
}

func (p *Plugin) removeUserSession(userID, channelID string) (channelState, channelState, error) {
	var currState channelState
	var prevState channelState
	err := p.kvSetAtomicChannelState(channelID, func(state *channelState) (*channelState, error) {
		if state == nil {
			return nil, fmt.Errorf("channel state is missing from store")
		}
		prevState = *state
		if state.Call == nil {
			return nil, fmt.Errorf("call state is missing from channel state")
		}

		if state.Call.ScreenSharingID == userID {
			state.Call.ScreenSharingID = ""
			if call := p.getCall(channelID); call != nil {
				call.setScreenSession(nil)
			}
		}

		delete(state.Call.Users, userID)

		if len(state.Call.Users) == 0 {
			state.Call = nil
			state.NodeID = ""
		}

		currState = *state
		return state, nil
	})

	return currState, prevState, err
}
