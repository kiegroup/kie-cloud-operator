package defaults

import (
	"fmt"
	"encoding/json"
	"os"
	"bytes"
	"io"
	"github.com/bmozaffa/rhpam-operator/configs"
)

func ConsoleEnvironmentDefaults() map[string]string {
	return overrideDefaults("configs/console-env.json")
}

func ServerEnvironmentDefaults() map[string]string {
	return overrideDefaults("configs/server-env.json")
}

func overrideDefaults(filename string) map[string]string {
	defaults := loadJsonMap("configs/common-env.json")
	configuration := loadJsonMap(filename)
	for key, value := range configuration {
		defaults[key] = value
	}
	return defaults
}

func loadJsonMap(filename string) map[string]string {
	bundle := configs.ConfigBundle
	file, e := bundle.Open(filename)
	if e != nil {
		fmt.Println("Failed to load %v", filename)
		return map[string]string{}
		os.Exit(1)
	}

	jsonMap := map[string]string{}
	json.Unmarshal(streamToByte(file), &jsonMap)
	return jsonMap
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
