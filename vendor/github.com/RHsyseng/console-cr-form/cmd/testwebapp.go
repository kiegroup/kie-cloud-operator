package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/go-openapi/spec"
	"github.com/sirupsen/logrus"
)

const defaultJSONForm = "test/examples/full-form.json"
const defaultJSONSchema = "test/examples/full-schema.json"
const envJSONForm = "JSON_FORM"
const envJSONSchema = "JSON_SCHEMA"

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("Starting test server. Using default JSON Form and Schema.")
	logrus.Info("Provide a different Form and Schema using JSON_FORM and JSON_SCHEMA env vars")

	config, err := web.NewConfiguration("", 8080, getSchema(), "app.kiegroup.org/v1", "KieApp", getForm(), callback)
	if err != nil {
		logrus.Fatalf("Failed to configure web server: %v", err)
	}
	if err := web.RunWebServer(config); err != nil {
		logrus.Fatalf("Failed to run web server: %v", err)
	}
}

func callback(yamlString string) error {
	logrus.Infof("Mock deploy yaml:\n%s", yamlString)
	if strings.Contains(yamlString, "fail-test") {
		errorMsg := `KieApp.app.kiegroup.org "fail-test" is invalid: []: Invalid value: map[string]interface {}{"status":map[string]interface {}
		{"deployments":map[string]interface {}{}, "conditions":interface {}(nil)}, "kind":"KieApp", "apiVersion":"app.kiegroup.org/v1",
		"metadata":map[string]interface {}{"namespace":"operator-ui", "generation":1, "uid":"29bc40ec-8388-11e9-9eb6-06daf81fbd22",
			"name":"fail-test", "creationTimestamp":"2019-05-31T09:40:45Z"}, "spec":map[string]interface {}{"auth":map[string]interface {}{},
			"upgrades":map[string]interface {}{}, "environment":"rhpam-trial", "imageRegistry":map[string]interface {}{"insecure":false},
			"objects":map[string]interface {}{"console":map[string]interface {}{"resources":map[string]interface {}{}},
			"servers":[]interface {}{map[string]interface {}{"build":map[string]interface {}{"gitSource":map[string]interface {}{},
			"webhooks":[]interface {}{map[string]interface {}{"type":"GitHub", "secret":"testsecret"}}}, "resources":map[string]interface {}{}}}},
			"commonConfig":map[string]interface {}{}}}: validation failure list:
			spec.objects.servers.build.gitSource.uri in body is required
			spec.objects.servers.build.gitSource.reference in body is required
			spec.objects.servers.build.kieServerContainerDeployment in body is required`
		logrus.Errorf("Return a canned error: %v", errorMsg)
		return fmt.Errorf(errorMsg)
	}
	return nil
}

func readJSONFile(envPath, defaultPath string) ([]byte, error) {
	filePath := getFilePath(envPath, defaultPath)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		logrus.Error("Unable to open file: ", err)
	}
	defer jsonFile.Close()
	return ioutil.ReadAll(jsonFile)
}

func getFilePath(env, defaultPath string) string {
	path := os.Getenv(env)
	if len(path) == 0 {
		return defaultPath
	}
	return path
}

func getForm() web.Form {
	byteValue, err := readJSONFile(envJSONForm, defaultJSONForm)
	if err != nil {
		logrus.Error("Unable to read file as byte array: ", err)
	}
	var form web.Form
	if err = json.Unmarshal(byteValue, &form); err != nil {
		logrus.Error("Error unmarshalling jsonForm: ", err)
	}
	return form
}

func getSchema() spec.Schema {
	byteValue, err := readJSONFile(envJSONSchema, defaultJSONSchema)
	if err != nil {
		logrus.Error("Unable to read file as byte array: ", err)
	}
	var schema spec.Schema
	if err = json.Unmarshal(byteValue, &schema); err != nil {
		logrus.Error("Error unmarshalling jsonSchema: ", err)
	}
	return schema
}
