package ui

//go:generate go run ../controller/kieapp/defaults/.packr/packr.go

import (
	"log"

	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
)

// Listen ...
func Listen() {
	config := &web.ConfigurationHolder{
		PortField:   8080,
		SchemaField: getSchema(),
		FormField:   getForm(),
	}
	if err := web.RunWebServer(config); err != nil {
		logrus.Errorf("Failed to run web server: %v", err)
	}
}

func getForm() web.Form {
	return web.Form{
		Pages: []web.Page{
			{
				Fields: []web.Field{
					{
						Label:    "Name",
						Default:  "rhpam-trial",
						Required: true,
						JSONPath: "$.metadata.name",
					},
					{
						Label:    "Environment",
						Default:  "rhpam-trial",
						Required: true,
						JSONPath: "$.spec.environment",
					},
				},
				Buttons: []web.Button{
					{
						Label:  "Cancel",
						Action: web.Cancel,
					},
					{
						Label:  "Deploy",
						Action: web.Submit,
					},
					{
						Label:  "Customize",
						Action: web.Next,
					},
				},
			},
		},
	}
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

func getSchema() spec.Schema {
	box := packr.New("CRD", "../../deploy/crds/")
	yamlByte, err := box.Find("kieapp.crd.yaml")
	if err != nil {
		log.Fatal(err)
		panic("Failed to retrieve crd, there must be an environment problem!")
	}
	crd := &CustomResourceDefinition{}
	err = yaml.Unmarshal(yamlByte, crd)
	if err != nil {
		panic("Failed to unmarshal static schema, there must be an environment problem!")
	}
	return crd.Spec.Validation.OpenAPIV3Schema
}
