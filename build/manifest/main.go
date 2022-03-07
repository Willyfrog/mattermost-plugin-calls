package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
)

const pluginIDGoFileTemplate = `// This file is automatically generated. Do not modify it manually.

package main

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v6/model"
)

var manifest model.Manifest
var wsActionPrefix string

const manifestStr = ` + "`" + `
%s
` + "`" + `

func init() {
	if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
		panic(err.Error())
	}
	wsActionPrefix = "custom_" + manifest.Id + "_"
}
`

const pluginIDJSFileTemplate = `// This file is automatically generated. Do not modify it manually.

const manifest = JSON.parse(` + "`" + `
%s
` + "`" + `);

export default manifest;
export const id = manifest.id;
export const version = manifest.version;
export const pluginId = manifest.id;
`

// These build-time vars are read from shell commands and populated in ../setup.mk
var (
	BuildHashShort  string
	BuildTagLatest  string
	BuildTagCurrent string
)

func main() {
	if len(os.Args) <= 1 {
		panic("no cmd specified")
	}

	manifest, err := findManifest()
	if err != nil {
		panic("failed to find manifest: " + err.Error())
	}

	cmd := os.Args[1]
	switch cmd {
	case "id":
		dumpPluginID(manifest)

	case "version":
		dumpPluginVersion(manifest)

	case "has_server":
		if manifest.HasServer() {
			fmt.Printf("true")
		}

	case "has_webapp":
		if manifest.HasWebapp() {
			fmt.Printf("true")
		}

	case "apply":
		if err := applyManifest(manifest); err != nil {
			panic("failed to apply manifest: " + err.Error())
		}

	case "dist":
		if err := distManifest(manifest); err != nil {
			panic("failed to write manifest to dist directory: " + err.Error())
		}

	default:
		panic("unrecognized command: " + cmd)
	}
}

func findManifest() (*model.Manifest, error) {
	_, manifestFilePath, err := model.FindManifest(".")
	if err != nil {
		return nil, errors.Wrap(err, "failed to find manifest in current working directory")
	}
	manifestFile, err := os.Open(manifestFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", manifestFilePath)
	}
	defer manifestFile.Close()

	// Re-decode the manifest, disallowing unknown fields. When we write the manifest back out,
	// we don't want to accidentally clobber anything we won't preserve.
	var manifest model.Manifest
	decoder := json.NewDecoder(manifestFile)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&manifest); err != nil {
		return nil, errors.Wrap(err, "failed to parse manifest")
	}

	// Update the manifest based on the state of the current commit
	// Use the first version we find (to prevent causing errors)
	var version string
	tags := strings.Fields(BuildTagCurrent)
	for _, t := range tags {
		if strings.HasPrefix(t, "v") {
			version = t
			break
		}
	}
	if version == "" {
		version = BuildTagLatest + "+" + BuildHashShort
	}
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}
	manifest.Version = version

	// Update the release notes url to point at the latest tag.
	manifest.ReleaseNotesURL = manifest.HomepageURL + "releases/tag/" + BuildTagLatest

	return &manifest, nil
}

// dumpPluginId writes the plugin id from the given manifest to standard out
func dumpPluginID(manifest *model.Manifest) {
	fmt.Printf("%s", manifest.Id)
}

// dumpPluginVersion writes the plugin version from the given manifest to standard out
func dumpPluginVersion(manifest *model.Manifest) {
	fmt.Printf("%s", manifest.Version)
}

// applyManifest propagates the plugin_id into the server and webapp folders, as necessary
func applyManifest(manifest *model.Manifest) error {
	if manifest.HasServer() {
		// generate JSON representation of Manifest.
		manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}
		manifestStr := string(manifestBytes)

		// write generated code to file by using Go file template.
		if err := ioutil.WriteFile(
			"server/manifest.go",
			[]byte(fmt.Sprintf(pluginIDGoFileTemplate, manifestStr)),
			0600,
		); err != nil {
			return errors.Wrap(err, "failed to write server/manifest.go")
		}
	}

	if manifest.HasWebapp() {
		// generate JSON representation of Manifest.
		// JSON is very similar and compatible with JS's object literals. so, what we do here
		// is actually JS code generation.
		manifestBytes, err := json.MarshalIndent(manifest, "", "    ")
		if err != nil {
			return err
		}
		manifestStr := string(manifestBytes)

		// write generated code to file by using JS file template.
		if err := ioutil.WriteFile(
			"webapp/src/manifest.ts",
			[]byte(fmt.Sprintf(pluginIDJSFileTemplate, manifestStr)),
			0600,
		); err != nil {
			return errors.Wrap(err, "failed to open webapp/src/manifest.ts")
		}
	}

	return nil
}

// distManifest writes the manifest file to the dist directory
func distManifest(manifest *model.Manifest) error {
	manifestBytes, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("dist/%s/plugin.json", manifest.Id), manifestBytes, 0600); err != nil {
		return errors.Wrap(err, "failed to write plugin.json")
	}

	return nil
}
