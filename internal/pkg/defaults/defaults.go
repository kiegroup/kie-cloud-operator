package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"

	"github.com/sirupsen/logrus"
)

func GetEnvironment(cr *opv1.App) (v1.Environment, []byte, error) {
	var env v1.Environment
	// default to '1' Kie Server
	if cr.Spec.NumKieServers == 0 {
		cr.Spec.NumKieServers = 1
	}

	//password := []byte("mykeystorepass")
	password := shared.GeneratePassword(8)

	// create go template
	template := v1.Template{
		ApplicationName:  cr.Name,
		KeyStorePassword: string(password),
	}
	envTemplate := opv1.EnvTemplate{
		Template: template,
	}
	for i := 0; i < cr.Spec.NumKieServers; i++ {
		envTemplate.ServerCount = append(envTemplate.ServerCount, template)
	}

	yamlBytes, err := loadYaml(fmt.Sprintf("envs/%s.yaml", cr.Spec.Environment), envTemplate)
	if err != nil {
		return env, nil, err
	}
	err = yaml.Unmarshal(yamlBytes, &env)
	if err != nil {
		logrus.Error(err)
	}

	return env, password, nil
}

func loadYaml(filename string, t opv1.EnvTemplate) ([]byte, error) {
	box := packr.NewBox("../../../config/app")

	if box.Has(filename) {
		// important to parse template first, before unmarshalling into object
		return parseTemplate(t, box.Bytes(filename)), nil
	}

	return nil, fmt.Errorf("%s does not exist, environment not deployed", filename)
}

func parseTemplate(e opv1.EnvTemplate, objBytes []byte) []byte {
	var b bytes.Buffer

	tmpl, err := template.New(e.ApplicationName).Parse(string(objBytes[:]))
	if err != nil {
		logrus.Error(err)
	}

	// template replacement
	err = tmpl.Execute(&b, e)
	if err != nil {
		logrus.Error(err)
	}

	return b.Bytes()
}

func GetConsoleObject() v1.CustomObject {
	object := v1.CustomObject{}
	yamlBytes, err := loadYaml("console.yaml", opv1.EnvTemplate{})
	if err != nil {
		logrus.Error(err)
	}
	yaml.Unmarshal(yamlBytes, &object)

	return object
}

func GetServerObject() v1.CustomObject {
	object := v1.CustomObject{}
	yamlBytes, err := loadYaml("server.yaml", opv1.EnvTemplate{})
	if err != nil {
		logrus.Error(err)
	}
	yaml.Unmarshal(yamlBytes, &object)

	return object
}
