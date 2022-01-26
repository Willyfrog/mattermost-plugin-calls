// This file is automatically generated. Do not modify it manually.

package main

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v6/model"
)

var manifest model.Manifest
var wsActionPrefix string

const manifestStr = `
{
  "id": "com.mattermost.calls",
  "name": "Calls",
  "description": "Integrates real-time voice communication in Mattermost",
  "version": "0.3.2-community",
  "min_server_version": "6.3.0",
  "server": {
    "executables": {
      "darwin-amd64": "server/dist/plugin-darwin-amd64",
      "darwin-arm64": "server/dist/plugin-darwin-arm64",
      "linux-amd64": "server/dist/plugin-linux-amd64",
      "linux-arm64": "server/dist/plugin-linux-arm64",
      "windows-amd64": "server/dist/plugin-darwin-amd64.exe"
    },
    "executable": ""
  },
  "webapp": {
    "bundle_path": "webapp/dist/main.js"
  },
  "settings_schema": {
    "header": "",
    "footer": "",
    "settings": [
      {
        "key": "ICEHostOverride",
        "display_name": "ICE Host Override",
        "type": "text",
        "help_text": "The IP (or hostname) to be used as the host ICE candidate. If empty, it defaults to resolving via STUN.",
        "placeholder": "",
        "default": ""
      },
      {
        "key": "UDPServerPort",
        "display_name": "RTC Server Port",
        "type": "number",
        "help_text": "The UDP port the RTC server will listen on.",
        "placeholder": "8443",
        "default": 8443
      },
      {
        "key": "ICEServers",
        "display_name": "ICE Servers",
        "type": "text",
        "help_text": "A comma separated list of ICE servers URLs (STUN/TURN) to use.",
        "placeholder": "stun:example.com:3478",
        "default": "stun:stun.l.google.com:19302,stun:global.stun.twilio.com:3478"
      },
      {
        "key": "AllowEnableCalls",
        "display_name": "Allow Enable Calls",
        "type": "bool",
        "help_text": "When set to true, it allows channel admins to enable calls in their channels. It also allows participants of DMs/GMs to enable calls.",
        "placeholder": "",
        "default": false
      }
    ]
  }
}
`

func init() {
	if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
		panic(err.Error())
	}
	wsActionPrefix = "custom_" + manifest.Id + "_"
}
