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
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func TestExampleCustomResources(t *testing.T) {
	schema := getSchema(t, api.SchemeGroupVersion.Version)
	box := packr.New("deploy/crs/v2", "../../../../deploy/crs/v2")
	assert.Greater(t, len(box.List()), 0)
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.Nil(t, err)
		yamlString = snippets(t, file, yamlString)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
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
		if strings.HasPrefix(missing.Path, "/status/conditions/lastTransitionTime") {
			// ...
		} else if strings.Contains(missing.Path, "/env/valueFrom/") {
			//The valueFrom is not expected to be used and is not fully defined TODO: verify
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

// for proper crd validation of snippet files... add required environment field to yaml
func snippets(t *testing.T, file, yamlStr string) string {
	if strings.Split(file, "/")[0] == "snippets" {
		jsonByte, err := yaml.YAMLToJSON([]byte(yamlStr))
		assert.Nil(t, err)
		jsonByte, err = sjson.SetBytes(jsonByte, "spec.environment", api.RhpamTrial)
		assert.Nil(t, err)
		yamlByte, err := yaml.JSONToYAML(jsonByte)
		assert.Nil(t, err)
		return string(yamlByte)
	}
	return yamlStr
}

func TestJvmCrd(t *testing.T) {
	crdYAML, _ := packr.New("deploy/crds", "../../../../deploy/crds").FindString("kieapp.crd.yaml")
	crdJSON, _ := yaml.YAMLToJSON([]byte(crdYAML))
	path := "spec.versions.#(name==" + api.SchemeGroupVersion.Version + ").schema.openAPIV3Schema.properties.spec.properties.objects.properties.servers.items.properties.jvm.properties"
	jvm := gjson.Get(string(crdJSON), path)

	testString(t, "javaOptsAppend", jvm)
	testInteger(t, "javaMaxMemRatio", jvm)
	testInteger(t, "javaInitialMemRatio", jvm)
	testInteger(t, "javaMaxInitialMem", jvm)
	testBoolean(t, "javaDiagnostics", jvm)
	testBoolean(t, "javaDebug", jvm)
	testInteger(t, "javaDebugPort", jvm)
	testInteger(t, "gcMinHeapFreeRatio", jvm)
	testInteger(t, "gcMaxHeapFreeRatio", jvm)
	testInteger(t, "gcTimeRatio", jvm)
	testInteger(t, "gcAdaptiveSizePolicyWeight", jvm)
	testString(t, "gcMaxMetaspaceSize", jvm)
	testString(t, "gcContainerOptions", jvm)
}

func testInteger(t *testing.T, name string, jvm gjson.Result) {
	assert.NotEmpty(t, jvm.Get(name+".description"))
	assert.Equal(t, "integer", jvm.Get(name+".type").String())
	assert.Equal(t, "int32", jvm.Get(name+".format").String())
}

func testString(t *testing.T, name string, jvm gjson.Result) {
	assert.NotEmpty(t, jvm.Get(name+".description"))
	assert.Equal(t, "string", jvm.Get(name+".type").String())
}

func testBoolean(t *testing.T, name string, jvm gjson.Result) {
	assert.NotEmpty(t, jvm.Get(name+".description"))
	assert.Equal(t, "boolean", jvm.Get(name+".type").String())
}
