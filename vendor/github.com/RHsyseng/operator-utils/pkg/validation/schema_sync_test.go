package validation

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type sampleApp struct {
	Spec   sampleAppSpec   `json:"spec,omitempty"`
	Status sampleAppStatus `json:"status,omitempty"`
}

type sampleAppSpec struct {
	SimpleText      string `json:"simpleText,omitempty"`
	secondAppObject `json:",inline"`
	IntPtr          *int32 `json:"intPtr,omitempty"`
	ObjArray        []env  `json:"envArray,omitempty"`
}

type secondAppObject struct {
	OtherText string `json:"otherText,omitempty"`
}

type sampleAppStatus struct {
	StatusText string `json:"statusText,omitempty"`
}

type env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

func TestSchemaStructComplaince(t *testing.T) {
	schema := getCompleteSchema(t)
	missingEntries := schema.GetMissingEntries(&sampleApp{})
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else {
			assert.Fail(t, "Discrepancy between CRD and Struct", "Missing or incorrect schema validation at %v, expected type %v", missing.Path, missing.Type)
		}
	}
}

func TestSchemaStructInlineJson(t *testing.T) {
	schema := getSchemaWithoutInline(t)
	missingEntries := schema.GetMissingEntries(&sampleApp{})
	assert.Len(t, missingEntries, 3, "Expect two status fields and one inline otherText field to be caught")
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else {
			assert.Equal(t, "/spec/otherText", missing.Path, "Other than status fields, expected to find /spec/otherText but instead found %s", missing.Path)
		}
	}
}

func TestSchemaStructIntPointer(t *testing.T) {
	schema := getSchemaWithoutIntPointer(t)
	missingEntries := schema.GetMissingEntries(&sampleApp{})
	assert.Len(t, missingEntries, 3, "Expect two status fields and one integer pointer field to be caught")
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else {
			assert.Equal(t, "/spec/intPtr", missing.Path, "Other than status fields, expected to find /spec/intPtr but instead found %s", missing.Path)
		}
	}
}

func TestSchemaStructSlice(t *testing.T) {
	schema := getSchemaWithoutSliceTypes(t)
	missingEntries := schema.GetMissingEntries(&sampleApp{})
	assert.Len(t, missingEntries, 4, "Expect two status fields and two sub-types of the slice to be caught")
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else if missing.Path == "/spec/envArray/name" {
			//Expected
		} else if missing.Path == "/spec/envArray/value" {
			//Expected
		} else {
			assert.Fail(t, "Unexpected validation failure", "Did not expect to fail with %s of type %s", missing.Path, missing.Type)
		}
	}
}

func TestSchemaFloat64(t *testing.T) {
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
          required:
          - number
          properties:
            number:
              type: number
              format: double
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)

	type myAppSpec struct {
		Number float64 `json:"number,omitempty"`
	}

	type myApp struct {
		Spec myAppSpec `json:"spec,omitempty"`
	}

	cr := myApp{
		Spec: myAppSpec{
			Number: float64(23),
		},
	}
	missingEntries := schema.GetMissingEntries(&cr)
	assert.Len(t, missingEntries, 0, "Expect no missing entries in CRD for this struct: %v", missingEntries)
}

func getCompleteSchema(t *testing.T) Schema {
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
            otherText:
              type: string
            intPtr:
              type: integer
            simpleObject:
              type: object
              properties:
                simpleField:
                  type: string
            envArray:
              type: array
              items:
                type: object
                properties:
                  name:
                    type: string
                  value:
                    type: string
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)
	return schema
}

func getSchemaWithoutInline(t *testing.T) Schema {
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
            intPtr:
              type: integer
            simpleObject:
              type: object
              properties:
                simpleField:
                  type: string
            envArray:
              type: array
              items:
                type: object
                properties:
                  name:
                    type: string
                  value:
                    type: string
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)
	return schema
}

func getSchemaWithoutIntPointer(t *testing.T) Schema {
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
            otherText:
              type: string
            simpleObject:
              type: object
              properties:
                simpleField:
                  type: string
            envArray:
              type: array
              items:
                type: object
                properties:
                  name:
                    type: string
                  value:
                    type: string
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)
	return schema
}

func getSchemaWithoutSliceTypes(t *testing.T) Schema {
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
            otherText:
              type: string
            intPtr:
              type: integer
            simpleObject:
              type: object
              properties:
                simpleField:
                  type: string
            envArray:
              type: array
`
	schema, err := New([]byte(schemaYaml))
	assert.NoError(t, err)
	return schema
}
