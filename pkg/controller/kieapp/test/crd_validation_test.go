package test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/gobuffalo/packr"
	"github.com/stretchr/testify/assert"
)

type CustomResourceDefinition struct {
	Spec CustomResourceDefinitionSpec `json:"spec,omitempty"`
}

type CustomResourceDefinitionSpec struct {
	Validation CustomResourceDefinitionValidation `json:"validation,omitempty"`
}

type CustomResourceDefinitionValidation struct {
	OpenAPIV3Schema spec.Schema `json:"openAPIV3Schema,omitempty"`
}

func TestSampleCustomResources(t *testing.T) {
	schema := getSchema(t)
	box := packr.NewBox("../../../../deploy/crs")
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, validate.AgainstSchema(schema, input, strfmt.Default), "File %v does not validate against the CRD schema", file)
	}
}

func TestTrialEnvMinimum(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v1
kind: KieApp
metadata:
  name: trial
spec:
  environment: rhpam-trial
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSchema(t)
	assert.NoError(t, validate.AgainstSchema(schema, input, strfmt.Default))

	deleteNestedMapEntry(input, "spec", "environment")
	assert.Error(t, validate.AgainstSchema(schema, input, strfmt.Default))
}

func TestSSO(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v1
kind: KieApp
metadata:
  name: trial
spec:
  environment: rhdm-trial
  auth:
    sso:
      url: https://rh-sso.example.com
      realm: rhpam
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSchema(t)
	assert.NoError(t, validate.AgainstSchema(schema, input, strfmt.Default))

	deleteNestedMapEntry(input, "spec", "auth", "sso", "realm")
	assert.Error(t, validate.AgainstSchema(schema, input, strfmt.Default))
}

func TestConsole(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v1
kind: KieApp
metadata:
  name: trial
spec:
  environment: rhpam-trial
  objects:
    console:
      env:
      - name: key1
        value: value1
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSchema(t)
	assert.NoError(t, validate.AgainstSchema(schema, input, strfmt.Default))

	deleteNestedMapEntry(input, "spec", "objects", "console", "env")
	//Validation commented out for now / OCP 3.11
	//assert.Error(t, validate.AgainstSchema(schema, input, strfmt.Default))

	deleteNestedMapEntry(input, "spec", "objects", "console")
	//Validation commented out for now / OCP 3.11
	//assert.Error(t, validate.AgainstSchema(schema, input, strfmt.Default))

	deleteNestedMapEntry(input, "spec", "objects")
	assert.NoError(t, validate.AgainstSchema(schema, input, strfmt.Default))
}

func deleteNestedMapEntry(object map[string]interface{}, keys ...string) {
	for index := 0; index < len(keys)-1; index++ {
		object = object[keys[index]].(map[string]interface{})
	}
	delete(object, keys[len(keys)-1])
}

func getSchema(t *testing.T) *spec.Schema {
	box := packr.NewBox("../../../../deploy/crds")
	crdFile := "kieapp.crd.yaml"
	assert.True(t, box.Has(crdFile))
	yamlString, err := box.FindString(crdFile)
	assert.NoError(t, err, "Error reading CRD yaml %v", yamlString)
	crd := &CustomResourceDefinition{}
	assert.NoError(t, yaml.Unmarshal([]byte(yamlString), crd))
	return &crd.Spec.Validation.OpenAPIV3Schema
}
