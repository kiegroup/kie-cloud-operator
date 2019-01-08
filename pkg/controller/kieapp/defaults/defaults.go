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
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var log = logs.GetLogger("kieapp.defaults")

func GetEnvironment(cr *v1.KieApp, service v1.PlatformService) (v1.Environment, error) {
	envTemplate := getEnvTemplate(cr)

	var common v1.Environment
	yamlBytes, err := loadYaml(service, "common.yaml", cr.Namespace, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &common)
	if err != nil {
		return v1.Environment{}, err
	}

	var env v1.Environment
	yamlBytes, err = loadYaml(service, fmt.Sprintf("envs/%s.yaml", cr.Spec.Environment), cr.Namespace, envTemplate)
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
	if cr.Spec.RhpamRegistry == (v1.KieAppRegistry{}) {
		cr.Spec.RhpamRegistry.Registry = logs.GetEnv("REGISTRY", constants.RhpamRegistry) // default to red hat registry
		cr.Spec.RhpamRegistry.Insecure = logs.GetBoolEnv("INSECURE")
	}

	// set default values for go template where not provided
	config := &cr.Spec.CommonConfig
	isTrialEnv := (cr.Spec.Environment == "trial")
	if len(config.Version) == 0 {
		pattern := regexp.MustCompile("[0-9]+")
		config.Version = strings.Join(pattern.FindAllString(constants.RhpamVersion, -1), "")
	}
	if len(config.ImageTag) == 0 {
		config.ImageTag = constants.ImageStreamTag
	}
	if len(config.ConsoleName) == 0 {
		if constants.MonitoringEnvs[cr.Spec.Environment] {
			config.ConsoleName = constants.RhpamcentrMonitoringServicePrefix
		} else {
			config.ConsoleName = constants.RhpamcentrServicePrefix
		}
	}
	if len(config.ConsoleImage) == 0 {
		if constants.MonitoringEnvs[cr.Spec.Environment] {
			config.ConsoleImage = constants.RhpamcentrMonitoringImageName
		} else {
			config.ConsoleImage = constants.RhpamcentrImageName
		}
	}

	setPassword(&config.KeyStorePassword, isTrialEnv)
	setPassword(&config.AdminPassword, isTrialEnv)
	setPassword(&config.ControllerPassword, isTrialEnv)
	setPassword(&config.ServerPassword, isTrialEnv)
	setPassword(&config.MavenPassword, isTrialEnv)

	crTemplate := v1.Template{
		CommonConfig:    config,
		ApplicationName: cr.Name,
	}
	envTemplate := v1.EnvTemplate{
		Template: crTemplate,
	}
	//For s2i KIE servers, the build configs determine the number and content of each KIE server group
	if len(cr.Spec.Objects.Builds) > 0 {
		cr.Spec.KieDeployments = len(cr.Spec.Objects.Builds)
		for _, build := range cr.Spec.Objects.Builds {
			buildTemplate := crTemplate.DeepCopy()
			buildTemplate.GitSource = build.GitSource
			buildTemplate.GitHubWebhookSecret = getWebhookSecret(v1.GitHubWebhook, build.Webhooks)
			buildTemplate.GenericWebhookSecret = getWebhookSecret(v1.GenericWebhook, build.Webhooks)
			buildTemplate.KieServerContainerDeployment = build.KieServerContainerDeployment
			envTemplate.ServerCount = append(envTemplate.ServerCount, *buildTemplate)
		}
	} else {
		for i := 0; i < cr.Spec.KieDeployments; i++ {
			envTemplate.ServerCount = append(envTemplate.ServerCount, crTemplate)
		}
	}

	return envTemplate
}

func setPassword(password *string, isTrialEnv bool) {
	if len(*password) != 0 {
		return
	}
	if isTrialEnv {
		*password = constants.DefaultPassword
	} else {
		*password = string(shared.GeneratePassword(8))
	}
}

func getWebhookSecret(webhookType v1.WebhookType, webhooks []v1.WebhookSecret) string {
	for _, webhook := range webhooks {
		if webhook.Type == webhookType {
			return webhook.Secret
		}
	}
	return string(shared.GeneratePassword(8))
}

// important to parse template first with this function, before unmarshalling into object
func loadYaml(service v1.PlatformService, filename, namespace string, e v1.EnvTemplate) ([]byte, error) {
	// use embedded files for tests, instead of ConfigMaps
	if service.IsMockService() {
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
	err := service.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, configMap)
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
