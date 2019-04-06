package validation

import (
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidSample(t *testing.T) {
	var inputYaml = `
apiVersion: app.example.com/v1
kind: SampleApp
metadata:
  name: test
spec:
  simpleText: value1
  simpleObject:
    simpleField: value2
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSampleSchema(t)
	assert.NoError(t, schema.Validate(input))
}

func TestValidSubset(t *testing.T) {
	var inputYaml = `
apiVersion: app.example.com/v1
kind: SampleApp
metadata:
  name: test
spec:
  simpleText: value
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSampleSchema(t)
	assert.NoError(t, schema.Validate(input))
}

func TestValidSuperset(t *testing.T) {
	var inputYaml = `
apiVersion: app.example.com/v1
kind: SampleApp
metadata:
  name: test
spec:
  simpleText: value1
  simpleObject:
    simpleField: value2
    simpleField2: value3
  simpleText2: value4
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSampleSchema(t)
	assert.NoError(t, schema.Validate(input))
}

func TestInValidSample(t *testing.T) {
	var inputYaml = `
apiVersion: app.example.com/v1
kind: SampleApp
metadata:
  name: test
spec:
  simpleText: value1
  simpleObject: value2
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSampleSchema(t)
	assert.Error(t, schema.Validate(input))
}

func getSampleSchema(t *testing.T) Schema {
	schemaYaml := `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: sample.app.example.com
spec:
  group: app.example.com
  names:
    kind: SampleApp
    listKind: SampleAppList
    plural: sampleapps
    singular: sampleapp
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      required:
        - spec
      properties:
        spec:
          type: object
          properties:
            simpleText:
              type: string
            simpleObject:
              type: object
              properties:
                simpleField:
                  type: string
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)
	return schema
}
