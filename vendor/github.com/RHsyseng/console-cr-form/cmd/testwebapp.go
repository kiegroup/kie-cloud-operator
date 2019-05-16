package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/go-openapi/spec"
	"github.com/sirupsen/logrus"
)

const defaultJSONForm = "test/examples/full-form.json"
const defaultJSONSchema = "test/examples/full-schema.json"
const envJSONForm = "JSON_FORM"
const envJSONSchema = "JSON_SCHEMA"

func main() {
	logrus.Info("Starting test server. Using default JSON Form and Schema.")
	logrus.Info("Provide a different Form and Schema using JSON_FORM and JSON_SCHEMA env vars")

	config, err := web.NewConfiguration("", 8080, getSchema(), "app.kiegroup.org/v1", "KieApp", getForm(), callback)
	if err != nil {
		logrus.Errorf("Failed to configure web server: %v", err)
	}
	if err := web.RunWebServer(config); err != nil {
		logrus.Errorf("Failed to run web server: %v", err)
	}
}

func callback(yamlString string) {
	logrus.Infof("Mock deploy yaml:\n%s", yamlString)
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

type CustomResourceDefinition struct {
	Spec CustomResourceDefinitionSpec `json:"spec,omitempty"`
}

type CustomResourceDefinitionSpec struct {
	Validation CustomResourceDefinitionValidation `json:"validation,omitempty"`
}

type CustomResourceDefinitionValidation struct {
	OpenAPIV3Schema spec.Schema `json:"openAPIV3Schema,omitempty"`
}
