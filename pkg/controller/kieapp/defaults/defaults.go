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
	"github.com/imdario/mergo"
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
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
	mergedEnv, err = mergeDB(service, cr, mergedEnv, envTemplate)
	if err != nil {
		return v1.Environment{}, err
	}
	return mergedEnv, nil
}

func mergeDB(service v1.PlatformService, cr *v1.KieApp, env v1.Environment, envTemplate v1.EnvTemplate) (v1.Environment, error) {
	dbEnvs := make(map[v1.DatabaseType]v1.Environment)
	for i := range env.Servers {
		kieServerSet := envTemplate.Servers[i]
		if kieServerSet.Database.Type == "" {
			continue
		}
		dbType := kieServerSet.Database.Type
		if _, loadedDB := dbEnvs[dbType]; !loadedDB {
			yamlBytes, err := loadYaml(service, fmt.Sprintf("dbs/%s.yaml", dbType), cr.Namespace, envTemplate)
			if err != nil {
				return v1.Environment{}, err
			}
			var dbEnv v1.Environment
			err = yaml.Unmarshal(yamlBytes, &dbEnv)
			if err != nil {
				return v1.Environment{}, err
			}
			dbEnvs[dbType] = dbEnv
		}
		dbServer, found := findCustomObjectByName(env.Servers[i], dbEnvs[dbType].Servers)
		if found {
			env.Servers[i] = mergeCustomObject(env.Servers[i], dbServer)
		}
	}
	return env, nil
}

func findCustomObjectByName(template v1.CustomObject, objects []v1.CustomObject) (v1.CustomObject, bool) {
	for i := range objects {
		if len(objects[i].DeploymentConfigs) == 0 || len(template.DeploymentConfigs) == 0 {
			return v1.CustomObject{}, false
		}
		if objects[i].DeploymentConfigs[0].ObjectMeta.Name == template.DeploymentConfigs[0].ObjectMeta.Name {
			return objects[i], true
		}
	}
	return v1.CustomObject{}, false
}

func getEnvTemplate(cr *v1.KieApp) (v1.EnvTemplate, error) {
	if cr.Spec.ImageRegistry == (v1.KieAppRegistry{}) {
		cr.Spec.ImageRegistry.Registry = logs.GetEnv("REGISTRY", constants.ImageRegistry) // default to red hat registry
		cr.Spec.ImageRegistry.Insecure = logs.GetBoolEnv("INSECURE")
	}

	// set default values for go template where not provided
	config := &cr.Spec.CommonConfig
	if config.ApplicationName == "" {
		config.ApplicationName = cr.Name
	}
	setAppConstants(&cr.Spec)
	isTrialEnv := strings.HasSuffix(string(cr.Spec.Environment), constants.TrialEnvSuffix)
	setPasswords(config, isTrialEnv)

	serversConfig, err := getServersConfig(cr, config)
	if err != nil {
		return v1.EnvTemplate{}, err
	}
	envTemplate := v1.EnvTemplate{
		CommonConfig: config,
		Console:      getConsoleTemplate(cr),
		Servers:      serversConfig,
		SmartRouter:  getSmartRouterTemplate(cr),
		Constants:    *getTemplateConstants(cr.Spec.Environment),
	}
	if err := configureAuth(cr, &envTemplate); err != nil {
		log.Error("unable to setup authentication: ", err)
		return envTemplate, err
	}

	return envTemplate, nil
}

func getTemplateConstants(env v1.EnvironmentType) *v1.TemplateConstants {
	c := constants.TemplateConstants.DeepCopy()
	if envConstants, found := constants.EnvironmentConstants[env]; found {
		c.Product = envConstants.App.Product
		c.MavenRepo = envConstants.App.MavenRepo
	}
	return c
}

func getConsoleTemplate(cr *v1.KieApp) v1.ConsoleTemplate {
	envConstants, hasEnv := constants.EnvironmentConstants[cr.Spec.Environment]
	template := v1.ConsoleTemplate{}
	if !hasEnv {
		return template
	}
	if cr.Spec.Objects.Console.KeystoreSecret == "" {
		template.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, strings.Join([]string{cr.Spec.CommonConfig.ApplicationName, "businesscentral"}, "-"))
	} else {
		template.KeystoreSecret = cr.Spec.Objects.Console.KeystoreSecret
	}
	// Set replicas
	envReplicas := v1.Replicas{}
	if hasEnv {
		envReplicas = envConstants.Replica.Console
	}
	replicas, denyScale := setReplicas(cr.Spec.Objects.Console.KieAppObject, envReplicas, hasEnv)
	if denyScale {
		cr.Spec.Objects.Console.Replicas = Pint32(replicas)
	}
	template.Replicas = replicas

	template.Name = envConstants.App.Prefix
	template.ImageName = envConstants.App.ImageName
	template.ProbePage = envConstants.App.ConsoleProbePage

	return template
}

func getSmartRouterTemplate(cr *v1.KieApp) v1.SmartRouterTemplate {
	envConstants, hasEnv := constants.EnvironmentConstants[cr.Spec.Environment]
	template := v1.SmartRouterTemplate{}
	if !hasEnv {
		return template
	}
	if cr.Spec.Objects.SmartRouter.KeystoreSecret == "" {
		template.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, strings.Join([]string{cr.Spec.CommonConfig.ApplicationName, "smartrouter"}, "-"))
	} else {
		template.KeystoreSecret = cr.Spec.Objects.SmartRouter.KeystoreSecret
	}

	// Set replicas
	envReplicas := v1.Replicas{}
	if hasEnv {
		envReplicas = envConstants.Replica.SmartRouter
	}
	replicas, denyScale := setReplicas(cr.Spec.Objects.SmartRouter, envReplicas, hasEnv)
	if denyScale {
		cr.Spec.Objects.SmartRouter.Replicas = Pint32(replicas)
	}
	template.Replicas = replicas

	return template
}

func setReplicas(object v1.KieAppObject, replicaConstant v1.Replicas, hasEnv bool) (replicas int32, denyScale bool) {
	if object.Replicas != nil {
		if hasEnv && replicaConstant.DenyScale && *object.Replicas != replicaConstant.Replicas {
			log.Warnf("scaling not allowed for this environment, setting to default of %d", replicaConstant.Replicas)
			return replicaConstant.Replicas, true
		}
		return *object.Replicas, false
	}
	if hasEnv {
		return replicaConstant.Replicas, false
	}
	log.Warnf("no replicas settings for this environment, defaulting to %d", replicas)
	return int32(1), denyScale
}

// serverSortBlanks moves blank names to the end
func serverSortBlanks(serverSets []v1.KieServerSet) []v1.KieServerSet {
	var newSets []v1.KieServerSet
	// servers with existing names should be placed in front
	for index := range serverSets {
		if serverSets[index].Name != "" {
			newSets = append(newSets, serverSets[index])
		}
	}
	// servers without names should be at the end
	for index := range serverSets {
		if serverSets[index].Name == "" {
			newSets = append(newSets, serverSets[index])
		}
	}
	if len(newSets) != len(serverSets) {
		log.Error("slice lengths aren't equal, returning server sets w/o blank names sorted")
		return serverSets
	}
	return newSets
}

// Returns the templates to use depending on whether the spec was defined with a common configuration
// or a specific one.
func getServersConfig(cr *v1.KieApp, commonConfig *v1.CommonConfig) ([]v1.ServerTemplate, error) {
	var servers []v1.ServerTemplate
	if len(cr.Spec.Objects.Servers) == 0 {
		cr.Spec.Objects.Servers = []v1.KieServerSet{{}}
	}
	cr.Spec.Objects.Servers = serverSortBlanks(cr.Spec.Objects.Servers)
	product := GetProduct(cr.Spec.Environment)
	usedNames := map[string]bool{}
	unsetNames := 0
	for index := range cr.Spec.Objects.Servers {
		serverSet := &cr.Spec.Objects.Servers[index]
		if serverSet.Deployments == nil {
			serverSet.Deployments = Pint(constants.DefaultKieDeployments)
		}
		if serverSet.Name == "" {
			for i := 0; i < len(cr.Spec.Objects.Servers); i++ {
				serverSetName := getKieSetName(cr.Spec.CommonConfig.ApplicationName, serverSet.Name, unsetNames)
				if !usedNames[serverSetName] {
					serverSet.Name = serverSetName
					break
				}
				unsetNames++
			}
		}
		for i := 0; i < *serverSet.Deployments; i++ {
			name := getKieDeploymentName(cr.Spec.CommonConfig.ApplicationName, serverSet.Name, unsetNames, i)
			if usedNames[name] {
				return []v1.ServerTemplate{}, fmt.Errorf("duplicate kieserver name %s", name)
			}
			usedNames[name] = true
			template := v1.ServerTemplate{
				KieName:        name,
				Build:          getBuildConfig(product, commonConfig, serverSet.Build),
				KeystoreSecret: serverSet.KeystoreSecret,
			}
			if serverSet.Build != nil {
				if *serverSet.Deployments > 1 {
					return []v1.ServerTemplate{}, fmt.Errorf("Cannot request %v deployments for a build", *serverSet.Deployments)
				}
				template.From = corev1.ObjectReference{
					Kind:      "ImageStreamTag",
					Name:      fmt.Sprintf("%s-kieserver:latest", commonConfig.ApplicationName),
					Namespace: "",
				}
			} else {
				template.From = getDefaultKieServerImage(product, commonConfig, serverSet.From)
			}

			// Set replicas
			envConstants, hasEnv := constants.EnvironmentConstants[cr.Spec.Environment]
			envReplicas := v1.Replicas{}
			if hasEnv {
				envReplicas = envConstants.Replica.Server
			}
			replicas, denyScale := setReplicas(serverSet.KieAppObject, envReplicas, hasEnv)
			if denyScale {
				serverSet.Replicas = Pint32(replicas)
			}
			template.Replicas = replicas

			dbConfig, err := getDatabaseConfig(cr.Spec.Environment, serverSet.Database)
			if err != nil {
				return servers, err
			}
			if dbConfig != nil {
				template.Database = *dbConfig
			}
			instanceTemplate := template.DeepCopy()
			if instanceTemplate.KeystoreSecret == "" {
				instanceTemplate.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, instanceTemplate.KieName)
			}
			servers = append(servers, *instanceTemplate)
		}
	}
	return servers, nil
}

// GetServerSet retrieves to correct ServerSet for processing and the DeploymentName
func GetServerSet(cr *v1.KieApp, requestedIndex int) (serverSet v1.KieServerSet, kieName string) {
	count := 0
	unnamedSets := 0
	for _, thisServerSet := range cr.Spec.Objects.Servers {
		for relativeIndex := 0; relativeIndex < *thisServerSet.Deployments; relativeIndex++ {
			if count == requestedIndex {
				serverSet = thisServerSet
				kieName = getKieDeploymentName(cr.Spec.CommonConfig.ApplicationName, serverSet.Name, unnamedSets, relativeIndex)
				return
			}
			count++
		}
		if thisServerSet.Name == "" {
			unnamedSets++
		}
	}
	return
}

// ConsolidateObjects construct all CustomObjects prior to creation
func ConsolidateObjects(env v1.Environment, cr *v1.KieApp) v1.Environment {
	env.Console = ConstructObject(env.Console, cr.Spec.Objects.Console.KieAppObject)
	env.SmartRouter = ConstructObject(env.SmartRouter, cr.Spec.Objects.SmartRouter)
	for index := range env.Servers {
		serverSet, _ := GetServerSet(cr, index)
		env.Servers[index] = ConstructObject(env.Servers[index], serverSet.KieAppObject)
	}
	return env
}

// ConstructObject returns an object after merging the environment object and the one defined in the CR
func ConstructObject(object v1.CustomObject, appObject v1.KieAppObject) v1.CustomObject {
	for dcIndex, dc := range object.DeploymentConfigs {
		for containerIndex, c := range dc.Spec.Template.Spec.Containers {
			c.Env = shared.EnvOverride(c.Env, appObject.Env)
			err := mergo.Merge(&c.Resources, appObject.Resources, mergo.WithOverride)
			if err != nil {
				log.Error("Error merging interfaces. ", err)
			}
			dc.Spec.Template.Spec.Containers[containerIndex] = c
		}
		object.DeploymentConfigs[dcIndex] = dc
	}
	return object
}

// getKieSetName aids in server indexing, depending on number of deployments and sets
func getKieSetName(applicationName string, setName string, arrayIdx int) string {
	return getKieDeploymentName(applicationName, setName, arrayIdx, 0)
}

func getKieDeploymentName(applicationName string, setName string, arrayIdx, deploymentsIdx int) string {
	name := setName
	if name == "" {
		name = fmt.Sprintf("%v-kieserver%v", applicationName, getKieSetIndex(arrayIdx, deploymentsIdx))
	} else {
		name = fmt.Sprintf("%v%v", setName, getKieSetIndex(0, deploymentsIdx))
	}
	return name
}

func getKieSetIndex(arrayIdx, deploymentsIdx int) string {
	var name bytes.Buffer
	if arrayIdx > 0 {
		name.WriteString(fmt.Sprintf("%d", arrayIdx+1))
	}
	if deploymentsIdx > 0 {
		name.WriteString(fmt.Sprintf("-%d", deploymentsIdx+1))
	}
	return name.String()
}

func getBuildConfig(product string, config *v1.CommonConfig, build *v1.KieAppBuildObject) v1.BuildTemplate {
	if build == nil {
		return v1.BuildTemplate{}
	}
	buildTemplate := v1.BuildTemplate{
		GitSource:                    build.GitSource,
		GitHubWebhookSecret:          getWebhookSecret(v1.GitHubWebhook, build.Webhooks),
		GenericWebhookSecret:         getWebhookSecret(v1.GenericWebhook, build.Webhooks),
		KieServerContainerDeployment: build.KieServerContainerDeployment,
		MavenMirrorURL:               build.MavenMirrorURL,
		ArtifactDir:                  build.ArtifactDir,
	}
	buildTemplate.From = getDefaultKieServerImage(product, config, build.From)
	return buildTemplate
}

func getDefaultKieServerImage(product string, config *v1.CommonConfig, from *corev1.ObjectReference) corev1.ObjectReference {
	if from != nil {
		return *from
	}
	imageName := fmt.Sprintf("%s%s-kieserver-openshift:%s", product, config.Version, constants.ImageStreamTag)
	return corev1.ObjectReference{
		Kind:      "ImageStreamTag",
		Name:      imageName,
		Namespace: constants.ImageStreamNamespace,
	}
}

func getDatabaseConfig(environment v1.EnvironmentType, database *v1.DatabaseObject) (*v1.DatabaseObject, error) {
	envConstants := constants.EnvironmentConstants[environment]
	if envConstants == nil {
		return nil, nil
	}
	defaultDB := envConstants.Database
	if database == nil {
		return defaultDB, nil
	}

	if database.Type == v1.DatabaseExternal && database.ExternalConfig == nil {
		return nil, fmt.Errorf("external database configuration is mandatory for external database type")
	}
	if database.Size == "" && defaultDB != nil {
		resultDB := *database.DeepCopy()
		resultDB.Size = defaultDB.Size
		return &resultDB, nil
	}
	return database, nil
}

func setPasswords(config *v1.CommonConfig, isTrialEnv bool) {
	passwords := []*string{
		&config.KeyStorePassword,
		&config.AdminPassword,
		&config.DBPassword,
		&config.AMQPassword,
		&config.AMQClusterPassword,
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
func loadYaml(service v1.PlatformService, filename, namespace string, env v1.EnvTemplate) ([]byte, error) {
	if _, _, useEmbedded := UseEmbeddedFiles(service); useEmbedded {
		box := packr.NewBox("../../../../config")
		if box.Has(filename) {
			yamlString, err := box.FindString(filename)
			if err != nil {
				return nil, err
			}
			return parseTemplate(env, yamlString), nil
		}
		return nil, fmt.Errorf("%s does not exist, '%s' KieApp not deployed", filename, env.ApplicationName)
	}

	cmName, file := convertToConfigMapName(filename)
	configMap := &corev1.ConfigMap{}
	err := service.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, configMap)
	if err != nil {
		return nil, fmt.Errorf("%s/%s ConfigMap not yet accessible, '%s' KieApp not deployed. Retrying... ", namespace, cmName, env.ApplicationName)
	}
	log.Debugf("Reconciling '%s' KieApp with %s from ConfigMap '%s'", env.ApplicationName, file, cmName)
	return parseTemplate(env, configMap.Data[file]), nil
}

func parseTemplate(env v1.EnvTemplate, objYaml string) []byte {
	var b bytes.Buffer

	tmpl, err := template.New(env.ApplicationName).Delims("[[", "]]").Parse(objYaml)
	if err != nil {
		log.Error("Error creating new Go template. ", err)
	}

	// template replacement
	err = tmpl.Execute(&b, env)
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
	var configMaps []corev1.ConfigMap
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
	if len(spec.CommonConfig.Version) == 0 {
		pattern := regexp.MustCompile("[0-9]+")
		spec.CommonConfig.Version = strings.Join(pattern.FindAllString(constants.ProductVersion, -1), "")
	}
	if len(spec.CommonConfig.ImageTag) == 0 {
		spec.CommonConfig.ImageTag = constants.ImageStreamTag
	}
}

// Pint returns a pointer to an integer
func Pint(i int) *int {
	return &i
}

// Pint32 returns a pointer to an integer
func Pint32(i int32) *int32 {
	return &i
}

func GetProduct(env v1.EnvironmentType) (product string) {
	envConstants := constants.EnvironmentConstants[env]
	if envConstants != nil {
		product = envConstants.App.Product
	}
	return
}
