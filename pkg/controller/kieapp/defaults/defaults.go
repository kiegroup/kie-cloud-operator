package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"github.com/kiegroup/kie-cloud-operator/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var log = logs.GetLogger("kieapp.defaults")

// GetEnvironment returns an Environment from merging the common config and the config
// related to the environment set in the KieApp definition
func GetEnvironment(cr *v1.KieApp, service v1.PlatformService) (v1.Environment, error) {
	envTemplate, err := getEnvTemplate(cr)
	if err != nil {
		return v1.Environment{}, err
	}

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

func getEnvTemplate(cr *v1.KieApp) (v1.EnvTemplate, error) {
	// default to '1' Kie DC
	if cr.Spec.KieDeployments == 0 {
		cr.Spec.KieDeployments = 1
	}
	if cr.Spec.ImageRegistry == (v1.KieAppRegistry{}) {
		cr.Spec.ImageRegistry.Registry = logs.GetEnv("REGISTRY", constants.ImageRegistry) // default to red hat registry
		cr.Spec.ImageRegistry.Insecure = logs.GetBoolEnv("INSECURE")
	}

	// set default values for go template where not provided
	config := &cr.Spec.CommonConfig
	setAppConstants(&cr.Spec)
	isTrialEnv := strings.HasSuffix(string(cr.Spec.Environment), constants.TrialEnvSuffix)
	setPasswords(config, isTrialEnv)

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
	if err := configureAuth(cr.Spec, &envTemplate); err != nil {
		log.Error("unable to setup authentication: ", err)
		return envTemplate, err
	}

	return envTemplate, nil
}

func setPasswords(config *v1.CommonConfig, isTrialEnv bool) {
	passwords := []*string{
		&config.KeyStorePassword,
		&config.AdminPassword,
		&config.ControllerPassword,
		&config.MavenPassword,
		&config.ServerPassword}

	for i := range passwords {
		if len(*passwords[i]) != 0 {
			continue
		}
		if isTrialEnv {
			*passwords[i] = constants.DefaultPassword
		} else {
			*passwords[i] = string(shared.GeneratePassword(8))
		}
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
	if _, _, useEmbedded := UseEmbeddedFiles(service); useEmbedded {
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

// ConfigMapsFromFile reads the files under the config folder and creates
// configmaps in the given namespace. It sets OwnerRef to operator deployment.
func ConfigMapsFromFile(myDep *appsv1.Deployment, ns string, scheme *runtime.Scheme) []corev1.ConfigMap {
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
				Namespace: ns,
				Annotations: map[string]string{
					v1.SchemeGroupVersion.Group: version.Version,
				},
			},
			Data: cmData,
		}

		cm.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		err := controllerutil.SetControllerReference(myDep, &cm, scheme)
		if err != nil {
			log.Error("Error setting controller reference. ", err)
		}
		for index := range cm.OwnerReferences {
			cm.OwnerReferences[index].BlockOwnerDeletion = nil
		}
		configMaps = append(configMaps, cm)
	}
	return configMaps
}

// UseEmbeddedFiles checks environment variables WATCH_NAMESPACE & OPERATOR_NAME
func UseEmbeddedFiles(service v1.PlatformService) (opName string, depNameSpace string, useEmbedded bool) {
	namespace := os.Getenv(constants.NameSpaceEnv)
	name := os.Getenv(constants.OpNameEnv)
	if service.IsMockService() || namespace == "" || name == "" {
		return name, namespace, true
	}
	return name, namespace, false
}

// setAppConstants sets the application-related constants to use in the template processing
func setAppConstants(spec *v1.KieAppSpec) {
	env := spec.Environment
	appConstants, hasEnv := constants.EnvironmentConstants[env]
	if !hasEnv {
		return
	}
	if len(spec.CommonConfig.Version) == 0 {
		pattern := regexp.MustCompile("[0-9]+")
		spec.CommonConfig.Version = strings.Join(pattern.FindAllString(constants.ProductVersion, -1), "")
	}
	if len(spec.CommonConfig.ImageTag) == 0 {
		spec.CommonConfig.ImageTag = constants.ImageStreamTag
	}
	if len(spec.CommonConfig.Product) == 0 {
		spec.CommonConfig.Product = appConstants.Product
	}
	if len(spec.CommonConfig.ConsoleName) == 0 {
		spec.CommonConfig.ConsoleName = appConstants.Prefix
	}
	if len(spec.CommonConfig.ConsoleImage) == 0 {
		spec.CommonConfig.ConsoleImage = appConstants.ImageName
	}
	if len(spec.CommonConfig.MavenRepo) == 0 {
		spec.CommonConfig.MavenRepo = appConstants.MavenRepo
	}
	if len(spec.CommonConfig.ConsoleProbePage) == 0 {
		spec.CommonConfig.ConsoleProbePage = appConstants.ConsoleProbePage
	}
}
