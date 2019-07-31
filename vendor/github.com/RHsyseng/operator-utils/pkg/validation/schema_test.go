package validation

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
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

func TestValidVersionedSuperset(t *testing.T) {
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

	schema := getSampleVersionedSchema(t, "v1")
	assert.NoError(t, schema.Validate(input))
}

func TestInValidVersionedSample(t *testing.T) {
	var inputYaml = `
apiVersion: app.example.com/v1beta1
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

	schema := getSampleVersionedSchema(t, "v1beta1")
	assert.Error(t, schema.Validate(input))
}

func TestMissingVersion(t *testing.T) {
	schema := getSampleVersionedSchema(t, "v1alpha1")
	assert.Empty(t, schema)
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

func getSampleVersionedSchema(t *testing.T, version string) Schema {
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
  versions:
    - name: v1
      schema:
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
    - name: v1beta1
      schema:
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
                      type: integer
`

	schema, err := NewVersioned([]byte(schemaYaml), version)
	if err != nil {
		assert.EqualError(t, err, fmt.Sprintf("no version %s detected in crd", version))
	}
	return schema
}
