package test

import (
	"strings"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/validation"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/stretchr/testify/assert"
)

func TestSampleCustomResources(t *testing.T) {
	schema := getSchema(t, api.SchemeGroupVersion.Version)
	box := packr.New("deploy/crs", "../../../../deploy/crs")
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
	}
}

func TestExampleCustomResources(t *testing.T) {
	for _, version := range getAPIVersions(t) {
		schema := getSchema(t, version)
		box := packr.New("deploy/examples"+version, "../../../../deploy/examples"+version)
		for _, file := range box.List() {
			yamlString, err := box.FindString(file)
			assert.NoError(t, err, "Error reading %v CR yaml", file)
			var input map[string]interface{}
			assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
			assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
		}
	}
}

func TestTrialEnvMinimum(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v2
kind: KieApp
metadata:
  name: trial
spec:
  environment: rhpam-trial
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSchema(t, api.SchemeGroupVersion.Version)
	assert.NoError(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "environment")
	assert.Error(t, schema.Validate(input))
}

func TestSSO(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v2
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

	schema := getSchema(t, api.SchemeGroupVersion.Version)
	assert.NoError(t, schema.Validate(input))

	deleteNestedMapEntry(input, "spec", "auth", "sso", "realm")
	assert.Error(t, schema.Validate(input))
}

func TestConsole(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v2
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

	schema := getSchema(t, api.SchemeGroupVersion.Version)
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

func TestCompleteCRD(t *testing.T) {
	schema := getSchema(t, api.SchemeGroupVersion.Version)
	missingEntries := schema.GetMissingEntries(&api.KieApp{})
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else if strings.Contains(missing.Path, "/env/valueFrom/") {
			//The valueFrom is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/uid") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/apiVersion") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/resourceVersion") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/fieldPath") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else {
			assert.Fail(t, "Discrepancy between CRD and Struct", "Missing or incorrect schema validation at %v, expected type %v", missing.Path, missing.Type)
		}
	}
}

func deleteNestedMapEntry(object map[string]interface{}, keys ...string) {
	for index := 0; index < len(keys)-1; index++ {
		object = object[keys[index]].(map[string]interface{})
	}
	delete(object, keys[len(keys)-1])
}

func getSchema(t *testing.T, version string) validation.Schema {
	box := packr.New("deploy/crds", "../../../../deploy/crds")
	crdFile := "kieapp.crd.yaml"
	assert.True(t, box.Has(crdFile))
	yamlString, err := box.FindString(crdFile)
	assert.NoError(t, err, "Error reading CRD yaml %v", yamlString)
	schema, err := validation.NewVersioned([]byte(yamlString), version)
	assert.NoError(t, err)
	return schema
}

func getAPIVersions(t *testing.T) (apiVersions []string) {
	isUnique := map[string]bool{}
	for _, configs := range constants.VersionConstants {
		if !isUnique[configs.APIVersion] {
			isUnique[configs.APIVersion] = true
			apiVersions = append(apiVersions, configs.APIVersion)
		}
	}
	assert.Contains(t, apiVersions, api.SchemeGroupVersion.Version)
	return apiVersions
}
