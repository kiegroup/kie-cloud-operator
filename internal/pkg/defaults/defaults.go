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
	"github.com/kiegroup/kie-cloud-operator/internal/constants"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"

	"github.com/sirupsen/logrus"
)

func GetEnvironment(cr *opv1.App) (opv1.Environment, []byte, error) {
	var env opv1.Environment
	// default to '1' Kie DC

	if cr.Spec.KieDeployments == 0 {
		cr.Spec.KieDeployments = 1
	}
	if cr.Spec.Version == "" {
		cr.Spec.Version = constants.RhpamVersion
	}
	if cr.Spec.ImageTag == "" {
		cr.Spec.ImageTag = constants.ImageStreamTag
	}

	password := shared.GeneratePassword(8)

	re := regexp.MustCompile("[0-9]+")
	// create go template
	template := Template{
		ApplicationName:    cr.Name,
		Version:            strings.Join(re.FindAllString(cr.Spec.Version, -1), ""),
		ImageTag:           cr.Spec.ImageTag,
		KeyStorePassword:   string(password),
		AdminPassword:      string(shared.GeneratePassword(8)),
		ControllerPassword: string(shared.GeneratePassword(8)),
		ServerPassword:     string(shared.GeneratePassword(8)),
		MavenPassword:      string(shared.GeneratePassword(8)),
	}
	envTemplate := EnvTemplate{
		Template: template,
	}
	for i := 0; i < cr.Spec.KieDeployments; i++ {
		envTemplate.ServerCount = append(envTemplate.ServerCount, template)
	}

	commonBytes, err := loadYaml("commonConfigs.yaml", envTemplate)
	if err != nil {
		return env, nil, err
	}
	var common opv1.AppSpec
	err = yaml.Unmarshal(commonBytes, &common)
	if err != nil {
		logrus.Error(err)
	}

	cr.Spec.Objects.Console.Env = shared.EnvOverride(common.Objects.Console.Env, cr.Spec.Objects.Console.Env)
	cr.Spec.Objects.Server.Env = shared.EnvOverride(common.Objects.Server.Env, cr.Spec.Objects.Server.Env)

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

func loadYaml(filename string, t EnvTemplate) ([]byte, error) {
	box := packr.NewBox("../../../config/app")

	if box.Has(filename) {
		// important to parse template first, before unmarshalling into object
		return parseTemplate(t, box.Bytes(filename)), nil
	}

	return nil, fmt.Errorf("%s does not exist, environment not deployed", filename)
}

func parseTemplate(e EnvTemplate, objBytes []byte) []byte {
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
