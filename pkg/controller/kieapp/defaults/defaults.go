package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var log = logs.GetLogger("kieapp.defaults")

func GetEnvironment(cr *v1.KieApp, client client.Client) (v1.Environment, error) {
	envTemplate := getEnvTemplate(cr)

	var common v1.Environment
	yamlBytes, err := loadYaml(client, "common.yaml", cr.Namespace, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &common)
	if err != nil {
		return v1.Environment{}, err
	}

	var env v1.Environment
	yamlBytes, err = loadYaml(client, fmt.Sprintf("envs/%s.yaml", cr.Spec.Environment), cr.Namespace, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &env)
	if err != nil {
		return v1.Environment{}, err
	}

	mergedEnv, err := merge(common, env)
	if err != nil {
		return v1.Environment{}, err
	}
	return mergedEnv, nil
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
	if cr.Spec.RhpamRegistry == (v1.KieAppRegistry{}) {
		cr.Spec.RhpamRegistry.Registry = logs.GetEnv("REGISTRY", constants.RhpamRegistry) // default to red hat registry
		cr.Spec.RhpamRegistry.Insecure = logs.GetBoolEnv("INSECURE")
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

// important to parse template first with this function, before unmarshalling into object
func loadYaml(client client.Client, filename, namespace string, e v1.EnvTemplate) ([]byte, error) {
	// use embedded files for tests, instead of ConfigMaps
	if reflect.DeepEqual(client, fake.NewFakeClient()) {
		box := packr.NewBox("../../../../config")
		if box.Has(filename) {
			yamlString, err := box.FindString(filename)
			if err != nil {
				return nil, err
			}
			return parseTemplate(e, yamlString), nil
		}
		return nil, fmt.Errorf("%s does not exist, '%s' KieApp not deployed", filename, e.ApplicationName)
	}

	cmName, file := convertToConfigMapName(filename)
	configMap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, configMap)
	if err != nil {
		return nil, fmt.Errorf("%s/%s ConfigMap not yet accessible, '%s' KieApp not deployed. Retrying... ", namespace, cmName, e.ApplicationName)
	}
	log.Debugf("Reconciling '%s' KieApp with %s from ConfigMap '%s'", e.ApplicationName, file, cmName)
	return parseTemplate(e, configMap.Data[file]), nil
}

func parseTemplate(e v1.EnvTemplate, objYaml string) []byte {
	var b bytes.Buffer

	tmpl, err := template.New(e.ApplicationName).Parse(objYaml)
	if err != nil {
		log.Error("Error creating new Go template. ", err)
	}

	// template replacement
	err = tmpl.Execute(&b, e)
	if err != nil {
		log.Error("Error applying Go template. ", err)
	}

	return b.Bytes()
}

func convertToConfigMapName(filename string) (configMapName, file string) {
	name := constants.ConfigMapPrefix
	result := strings.Split(filename, "/")
	if len(result) > 1 {
		for i := 0; i < len(result)-1; i++ {
			name = strings.Join([]string{name, result[i]}, "-")
		}
	}
	return name, result[len(result)-1]
}

func ConfigMapsFromFile(namespace string) []corev1.ConfigMap {
	box := packr.NewBox("../../../../config")
	cmList := map[string][]map[string]string{}
	for _, filename := range box.List() {
		s, err := box.FindString(filename)
		if err != nil {
			log.Error("Error finding file with packr. ", err)
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
