package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	corev1 "k8s.io/api/core/v1"

	"github.com/sirupsen/logrus"
)

// GetEnvironment Loads the commonConfigs.yaml file and then overrides the values
// with the provided env yaml file. e.g. envs/production-lite.yaml
func GetEnvironment(cr *v1.KieApp) (v1.Environment, v1.KieAppSpec, error) {
	var env v1.Environment
	envTemplate := getEnvTemplate(cr)

	commonBytes, err := loadYaml("commonConfigs.yaml", envTemplate)
	if err != nil {
		return env, v1.KieAppSpec{}, err
	}
	var common v1.KieAppSpec
	err = yaml.Unmarshal(commonBytes, &common)
	if err != nil {
		logrus.Error(err)
	}

	yamlBytes, err := loadYaml(fmt.Sprintf("envs/%s.yaml", cr.Spec.Environment), envTemplate)
	if err != nil {
		return env, common, err
	}
	err = yaml.Unmarshal(yamlBytes, &env)
	if err != nil {
		logrus.Error(err)
	}

	return env, common, nil
}

func getEnvTemplate(cr *v1.KieApp) v1.EnvTemplate {
	// default to '1' Kie DC
	if cr.Spec.KieDeployments == 0 {
		cr.Spec.KieDeployments = 1
	}
	if len(cr.Spec.Objects.Console.Env) == 0 {
		cr.Spec.Objects.Console.Env = []corev1.EnvVar{{Name: "empty"}}
	}
	if len(cr.Spec.Objects.Server.Env) == 0 {
		cr.Spec.Objects.Server.Env = []corev1.EnvVar{{Name: "empty"}}
	}

	pattern := regexp.MustCompile("[0-9]+")
	// create go template if does not exist
	if cr.Spec.Template.ApplicationName == "" {
		cr.Spec.Template = v1.Template{
			ApplicationName:    cr.Name,
			Version:            strings.Join(pattern.FindAllString(constants.RhpamVersion, -1), ""),
			ImageTag:           constants.ImageStreamTag,
			KeyStorePassword:   string(shared.GeneratePassword(8)),
			AdminPassword:      string(shared.GeneratePassword(8)),
			ControllerPassword: string(shared.GeneratePassword(8)),
			ServerPassword:     string(shared.GeneratePassword(8)),
			MavenPassword:      string(shared.GeneratePassword(8)),
		}
	}
	envTemplate := v1.EnvTemplate{
		Template: cr.Spec.Template,
	}
	for i := 0; i < cr.Spec.KieDeployments; i++ {
		envTemplate.ServerCount = append(envTemplate.ServerCount, cr.Spec.Template)
	}

	return envTemplate
}

func loadYaml(filename string, t v1.EnvTemplate) ([]byte, error) {
	box := packr.NewBox("../../../../config")

	if box.Has(filename) {
		// important to parse template first, before unmarshalling into object
		bytes, err := box.Find(filename)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		return parseTemplate(t, bytes), nil
	}

	return nil, fmt.Errorf("%s does not exist, environment not deployed", filename)
}

func parseTemplate(e v1.EnvTemplate, objBytes []byte) []byte {
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
