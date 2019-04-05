package test

import (
	"github.com/RHsyseng/operator-utils/pkg/validation"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/stretchr/testify/assert"
)

func TestSampleCustomResources(t *testing.T) {
	schema := getSchema(t)
	box := packr.NewBox("../../../../deploy/crs")
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
	}
}

func TestExampleCustomResources(t *testing.T) {
	schema := getSchema(t)
	box := packr.NewBox("../../../../deploy/examples")
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
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
	assert.NoError(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "environment")
	assert.Error(t, schema.Validate(input))
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
	assert.NoError(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "auth", "sso", "realm")
	assert.Error(t, schema.Validate(input))
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
	assert.NoError(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "objects", "console", "env")
	//Validation commented out for now / OCP 3.11
	//assert.Error(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "objects", "console")
	//Validation commented out for now / OCP 3.11
	//assert.Error(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "objects")
	assert.NoError(t, schema.Validate(input))
}

func deleteNestedMapEntry(object map[string]interface{}, keys ...string) {
	for index := 0; index < len(keys)-1; index++ {
		object = object[keys[index]].(map[string]interface{})
	}
	delete(object, keys[len(keys)-1])
}

func getSchema(t *testing.T) validation.Schema {
	box := packr.NewBox("../../../../deploy/crds")
	crdFile := "kieapp.crd.yaml"
	assert.True(t, box.Has(crdFile))
	yamlString, err := box.FindString(crdFile)
	assert.NoError(t, err, "Error reading CRD yaml %v", yamlString)
	schema, err := validation.New([]byte(yamlString))
	assert.NoError(t, err)
	return schema
}
