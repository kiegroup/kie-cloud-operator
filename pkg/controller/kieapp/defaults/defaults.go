package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"bytes"
	"context"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/sirupsen/logrus"
)

func GetLiteEnvironment(cr *v1.KieApp) (v1.Environment, error) {
	envTemplate := getEnvTemplate(cr)

	var servers v1.CustomObject
	yamlBytes, err := loadYaml(fake.NewFakeClient(), "common/server.yaml", cr.Namespace, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &servers)
	if err != nil {
		return v1.Environment{}, err
	}

	var console v1.CustomObject
	yamlBytes, err = loadYaml(fake.NewFakeClient(), "common/console.yaml", cr.Namespace, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &console)
	if err != nil {
		return v1.Environment{}, err
	}

	var env v1.Environment
	yamlBytes, err = loadYaml(fake.NewFakeClient(), fmt.Sprintf("envs/%s-lite.yaml", cr.Spec.Environment), cr.Namespace, envTemplate)
	logrus.Infof("Trial env is %v", string(yamlBytes))
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &env)
	if err != nil {
		return v1.Environment{}, err
	}

	merge(&console, &env.Console)
	env.Console = console
	for index := range env.Servers {
		merge(&servers, &env.Servers[index])
		env.Servers[index] = servers
	}

	return env, nil
}

// GetEnvironment Loads the commonConfigs.yaml file and then overrides the values
// with the provided env yaml file. e.g. envs/production-lite.yaml
func GetEnvironment(cr *v1.KieApp, client client.Client) (v1.Environment, v1.KieAppSpec, error) {
	var env v1.Environment
	envTemplate := getEnvTemplate(cr)

	commonBytes, err := loadYaml(client, "commonConfigs.yaml", cr.Namespace, envTemplate)
	if err != nil {
		return env, v1.KieAppSpec{}, err
	}
	var common v1.KieAppSpec
	err = yaml.Unmarshal(commonBytes, &common)
	if err != nil {
		logrus.Error(err)
	}

	yamlBytes, err := loadYaml(client, fmt.Sprintf("envs/%s.yaml", cr.Spec.Environment), cr.Namespace, envTemplate)
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

func loadYaml(client client.Client, filename, namespace string, t v1.EnvTemplate) ([]byte, error) {
	cmName, file := convertToConfigMapName(filename)
	logrus.Debugf("Loading contents from %s in ConfigMap '%s'", file, cmName)
	configMap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, configMap)
	if err != nil {
		logrus.Warnf("'%s - %s' ConfigMap does not exist, using embedded '%s'", cmName, file, filename)
		box := packr.NewBox("../../../../config")
		if box.Has(filename) {
			// important to parse template first, before unmarshalling into object
			yamlString, err := box.FindString(filename)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			return parseTemplate(t, yamlString), nil
		}
		return nil, fmt.Errorf("%s does not exist, environment not deployed", filename)
	}
	return parseTemplate(t, configMap.Data[file]), nil
}

func parseTemplate(e v1.EnvTemplate, objYaml string) []byte {
	var b bytes.Buffer

	tmpl, err := template.New(e.ApplicationName).Parse(objYaml)
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

func convertToConfigMapName(filename string) (configMapName, file string) {
	name := constants.ConfigMapPrefix
	result := strings.Split(filename, "/")
	if len(result) > 1 {
		name = strings.Join([]string{name, result[0]}, "-")
	}
	return name, result[len(result)-1]
}

func ConfigMapsFromFile(namespace string) []corev1.ConfigMap {
	box := packr.NewBox("../../../../config")
	cmList := map[string][]map[string]string{}
	for _, filename := range box.List() {
		s, err := box.FindString(filename)
		if err != nil {
			logrus.Error(err)
		}
		cmData := map[string]string{}
		cmName, file := convertToConfigMapName(filename)
		cmData[file] = s
		cmList[cmName] = append(cmList[cmName], cmData)
	}
	configMaps := []corev1.ConfigMap{}
	for cmName, dataSlice := range cmList {
		cmData := map[string]string{}
		for _, dataList := range dataSlice {
			for name, data := range dataList {
				cmData[name] = data
			}
		}
		cm := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: namespace,
			},
			Data: cmData,
		}
		configMaps = append(configMaps, cm)
	}
	return configMaps
}
