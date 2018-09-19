package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"encoding/json"

	"github.com/gobuffalo/packr"
)

func ConsoleEnvironmentDefaults() map[string]string {
	return overrideDefaults("console-env.json")
}

func ServerEnvironmentDefaults() map[string]string {
	return overrideDefaults("server-env.json")
}

func overrideDefaults(filename string) map[string]string {
	defaults := loadJsonMap("common-env.json")
	configuration := loadJsonMap(filename)
	for key, value := range configuration {
		defaults[key] = value
	}
	return defaults
}

func loadJsonMap(filename string) map[string]string {
	box := packr.NewBox("../../../config/app")
	jsonMap := make(map[string]string)
	json.Unmarshal(box.Bytes(filename), &jsonMap)
	return jsonMap
}
