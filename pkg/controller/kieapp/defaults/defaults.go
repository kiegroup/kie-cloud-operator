package defaults

//go:generate go run -mod=vendor .packr/packr.go

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	"github.com/imdario/mergo"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"github.com/kiegroup/kie-cloud-operator/version"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var log = logs.GetLogger("kieapp.defaults")

// GetEnvironment returns an Environment from merging the common config and the config
// related to the environment set in the KieApp definition
func GetEnvironment(cr *api.KieApp, service kubernetes.PlatformService) (api.Environment, error) {
	minor, micro, err := checkProductUpgrade(cr)
	if err != nil {
		return api.Environment{}, err
	}
	// handle upgrade logic from here
	cMajor, _, _ := GetMajorMinorMicro(cr.Status.Applied.Version)
	lMajor, _, _ := GetMajorMinorMicro(constants.CurrentVersion)
	minorVersion := GetMinorImageVersion(cr.Status.Applied.Version)
	latestMinorVersion := GetMinorImageVersion(constants.CurrentVersion)
	if (micro && minorVersion == latestMinorVersion) ||
		(minor && minorVersion != latestMinorVersion && cMajor == lMajor) {
		if err := getConfigVersionDiffs(cr.Status.Applied.Version, constants.CurrentVersion, service); err != nil {
			return api.Environment{}, err
		}
		// reset current annotations and update CR to use latest product version
		cr.SetAnnotations(map[string]string{})
		cr.Status.Applied.Version = constants.CurrentVersion
		cr.Spec.Version = ""
	}
	envTemplate, err := getEnvTemplate(cr)
	if err != nil {
		return api.Environment{}, err
	}

	var common api.Environment
	yamlBytes, err := loadYaml(service, "common.yaml", cr.Status.Applied.Version, cr.Namespace, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &common)
	if err != nil {
		return api.Environment{}, err
	}
	var env api.Environment
	yamlBytes, err = loadYaml(service, fmt.Sprintf("envs/%s.yaml", cr.Status.Applied.Environment), cr.Status.Applied.Version, cr.Namespace, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &env)
	if err != nil {
		return api.Environment{}, err
	}
	if cr.Status.Applied.Objects.SmartRouter == nil {
		env.SmartRouter.Omit = true
	}

	if cr.Status.Applied.Environment == api.RhdmProductionImmutable || (isImmutable(cr) && cr.Status.Applied.Objects.Console == nil) {
		env.Console.Omit = true
	}

	mergedEnv, err := merge(common, env)
	if err != nil {
		return api.Environment{}, err
	}

	// if bc monitoring is not set, there's no need to set the environment variables below
	if mergedEnv.Console.Omit {
		// cleaning console object
		mergedEnv.Console = api.CustomObject{Omit: true}
	}

	if mergedEnv.SmartRouter.Omit {
		// remove router env vars from kieserver DCs
		for _, server := range mergedEnv.Servers {
			for _, dc := range server.DeploymentConfigs {
				newSlice := []corev1.EnvVar{}
				for _, envvar := range dc.Spec.Template.Spec.Containers[0].Env {
					if envvar.Name != "KIE_SERVER_ROUTER_SERVICE" && envvar.Name != "KIE_SERVER_ROUTER_PORT" && envvar.Name != "KIE_SERVER_ROUTER_PROTOCOL" {
						newSlice = append(newSlice, envvar)
					}
				}
				dc.Spec.Template.Spec.Containers[0].Env = newSlice
			}
		}
	}

	mergedEnv, err = mergeDB(service, cr, mergedEnv, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	mergedEnv, err = mergeJms(service, cr, mergedEnv, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	mergedEnv, err = mergeProcessMigration(service, cr, mergedEnv, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	mergedEnv, err = mergeDBDeployment(service, cr, mergedEnv, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	setProductLabels(cr, &mergedEnv)
	return mergedEnv, nil
}

func setProductLabels(cr *api.KieApp, env *api.Environment) {
	setObjectLabels(cr, &env.Console)
	for index := range env.Servers {
		setObjectLabels(cr, &env.Servers[index])
	}
	setObjectLabels(cr, &env.SmartRouter)

	setObjectLabels(cr, &env.ProcessMigration)
}

func getSubComponentTypeByImageName(imageName string) string {
	retValue := "application"
	if imageName == constants.RhpamSmartRouterImageName || imageName == constants.RhpamControllerImageName ||
		imageName == constants.RhdmSmartRouterImageName || imageName == constants.RhdmControllerImageName {

		retValue = "infrastructure"
	}
	return retValue
}

func setObjectLabels(cr *api.KieApp, object *api.CustomObject) {
	for index, obj := range object.DeploymentConfigs {
		subcomponent, _, _ := GetImage(object.DeploymentConfigs[index].Spec.Template.Spec.Containers[0].Image)
		object.DeploymentConfigs[index].Spec.Template.Labels = setLabels(cr, obj.Spec.Template.Labels, subcomponent, getSubComponentTypeByImageName(subcomponent))
	}
	for index, obj := range object.StatefulSets {
		subcomponent, _, _ := GetImage(object.StatefulSets[index].Spec.Template.Spec.Containers[0].Image)
		object.StatefulSets[index].Spec.Template.Labels = setLabels(cr, obj.Spec.Template.Labels, subcomponent, getSubComponentTypeByImageName(subcomponent))
	}

}

func setLabels(cr *api.KieApp, labels map[string]string, subcomponent string, subcomponentType string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[constants.LabelRHproductName] = constants.ProductName
	labels[constants.LabelRHproductVersion] = cr.Status.Applied.Version
	labels[constants.LabelRHcomponentName] = "PAM"
	labels[constants.LabelRHsubcomponentName] = subcomponent
	labels[constants.LabelRHcomponentVersion] = cr.Status.Applied.Version
	labels[constants.LabelRHsubcomponentType] = subcomponentType
	labels[constants.LabelRHcompany] = "Red_Hat"
	return labels
}

func mergeDB(service kubernetes.PlatformService, cr *api.KieApp, env api.Environment, envTemplate api.EnvTemplate) (api.Environment, error) {
	dbEnvs := make(map[api.DatabaseType]api.Environment)
	for i := range env.Servers {
		kieServerSet := envTemplate.Servers[i]
		if kieServerSet.Database.Type == "" {
			continue
		}
		dbType := kieServerSet.Database.Type
		if isGE78(cr) {
			if err := loadDBYamls(service, cr, envTemplate, "dbs/servers/%s.yaml", dbType, dbEnvs); err != nil {
				return api.Environment{}, err
			}
		} else if _, loadedDB := dbEnvs[dbType]; !loadedDB {
			yamlBytes, err := loadYaml(service, fmt.Sprintf("dbs/%s.yaml", dbType), cr.Spec.Version, cr.Namespace, envTemplate)
			if err != nil {
				return api.Environment{}, err
			}
			var dbEnv api.Environment
			err = yaml.Unmarshal(yamlBytes, &dbEnv)
			if err != nil {
				return api.Environment{}, err
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

func mergeJms(service kubernetes.PlatformService, cr *api.KieApp, env api.Environment, envTemplate api.EnvTemplate) (api.Environment, error) {
	var jmsEnv api.Environment
	for i := range env.Servers {
		kieServerSet := envTemplate.Servers[i]
		if kieServerSet.Jms.EnableIntegration {
			yamlBytes, err := loadYaml(service, fmt.Sprintf("jms/activemq-jms-config.yaml"), cr.Status.Applied.Version, cr.Namespace, envTemplate)
			if err != nil {
				return api.Environment{}, err
			}
			err = yaml.Unmarshal(yamlBytes, &jmsEnv)
			if err != nil {
				return api.Environment{}, err
			}
		}
		jmsServer, found := findCustomObjectByName(env.Servers[i], jmsEnv.Servers)
		if found {
			env.Servers[i] = mergeCustomObject(env.Servers[i], jmsServer)
		}
	}
	return env, nil
}

func findCustomObjectByName(template api.CustomObject, objects []api.CustomObject) (api.CustomObject, bool) {
	for i := range objects {
		if len(objects[i].DeploymentConfigs) == 0 || len(template.DeploymentConfigs) == 0 {
			return api.CustomObject{}, false
		}
		if objects[i].DeploymentConfigs[0].ObjectMeta.Name == template.DeploymentConfigs[0].ObjectMeta.Name {
			return objects[i], true
		}
	}
	return api.CustomObject{}, false
}

func getEnvTemplate(cr *api.KieApp) (envTemplate api.EnvTemplate, err error) {

	SetDefaults(cr)
	serversConfig, err := getServersConfig(cr)
	if err != nil {
		return envTemplate, err
	}
	envTemplate = api.EnvTemplate{
		Console:     getConsoleTemplate(cr),
		Servers:     serversConfig,
		SmartRouter: getSmartRouterTemplate(cr),
		Constants:   *getTemplateConstants(cr),
	}
	processMigrationConfig, err := getProcessMigrationTemplate(cr, serversConfig)
	if err != nil {
		return envTemplate, err
	}
	if processMigrationConfig != nil {
		envTemplate.ProcessMigration = *processMigrationConfig
	}
	envTemplate.Databases = getDatabaseDeploymentTemplate(cr, serversConfig, processMigrationConfig)
	envTemplate.CommonConfig = &cr.Status.Applied.CommonConfig
	if cr.Status.Applied.Auth != nil {
		if err := configureAuth(cr, &envTemplate); err != nil {
			log.Error("unable to setup authentication: ", err)
			return envTemplate, err
		}
	}
	return envTemplate, nil
}

func getTemplateConstants(cr *api.KieApp) *api.TemplateConstants {
	c := constants.TemplateConstants.DeepCopy()
	c.Major, c.Minor, c.Micro = GetMajorMinorMicro(cr.Status.Applied.Version)
	if envConstants, found := constants.EnvironmentConstants[cr.Status.Applied.Environment]; found {
		c.Product = envConstants.App.Product
		c.MavenRepo = envConstants.App.MavenRepo
	}
	if versionConstants, found := constants.VersionConstants[cr.Status.Applied.Version]; found {
		c.BrokerImageContext = versionConstants.BrokerImageContext
		c.BrokerImage = versionConstants.BrokerImage
		c.BrokerImageTag = versionConstants.BrokerImageTag
		c.DatagridImageContext = versionConstants.DatagridImageContext
		c.DatagridImage = versionConstants.DatagridImage
		c.DatagridImageTag = versionConstants.DatagridImageTag

		c.OseCliImageURL = versionConstants.OseCliImageURL
		c.MySQLImageURL = versionConstants.MySQLImageURL
		c.PostgreSQLImageURL = versionConstants.PostgreSQLImageURL
		c.DatagridImageURL = versionConstants.DatagridImageURL
		c.BrokerImageURL = versionConstants.BrokerImageURL
	}
	if val, exists := os.LookupEnv(constants.OseCliVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
		c.OseCliImageURL = val
	}
	if val, exists := os.LookupEnv(constants.MySQLVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
		c.MySQLImageURL = val
	}
	if val, exists := os.LookupEnv(constants.PostgreSQLVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
		c.PostgreSQLImageURL = val
	}
	if val, exists := os.LookupEnv(constants.DatagridVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
		c.DatagridImageURL = val
	}
	if val, exists := os.LookupEnv(constants.BrokerVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
		c.BrokerImageURL = val
	}
	return c
}

func getConsoleTemplate(cr *api.KieApp) api.ConsoleTemplate {
	template := api.ConsoleTemplate{}
	if cr.Status.Applied.Objects.Console != nil {
		envConstants, hasEnv := constants.EnvironmentConstants[cr.Status.Applied.Environment]

		// Set replicas
		if !hasEnv {
			return template
		}
		replicas, denyScale := setReplicas(cr.Status.Applied.Objects.Console.Replicas, envConstants.Replica.Console, hasEnv)
		if denyScale || cr.Status.Applied.Objects.Console.Replicas == nil {
			cr.Status.Applied.Objects.Console.Replicas = Pint32(replicas)
		}
		template.Replicas = *cr.Status.Applied.Objects.Console.Replicas
		template.Name = envConstants.App.Prefix
		cMajor, _, _ := GetMajorMinorMicro(cr.Status.Applied.Version)
		template.ImageURL = constants.ImageRegistry + "/" + envConstants.App.Product + "-" + cMajor + "/" + envConstants.App.Product + "-" + envConstants.App.ImageName + constants.RhelVersion + ":" + cr.Status.Applied.Version
		template.KeystoreSecret = cr.Status.Applied.Objects.Console.KeystoreSecret
		if template.KeystoreSecret == "" {
			template.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, strings.Join([]string{cr.Status.Applied.CommonConfig.ApplicationName, "businesscentral"}, "-"))
		}

		template.StorageClassName = cr.Status.Applied.Objects.Console.StorageClassName
		if !cr.Status.Applied.UseImageTags {
			if val, exists := os.LookupEnv(envConstants.App.ImageVar + cr.Status.Applied.Version); exists {
				template.ImageURL = val
			}
			template.OmitImageStream = true
		}
		template.Image, template.ImageTag, template.ImageContext = GetImage(template.ImageURL)
		if cr.Status.Applied.Objects.Console.Image != "" {
			template.Image = cr.Status.Applied.Objects.Console.Image
			template.ImageURL = template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.Console.ImageTag != "" {
			template.ImageTag = cr.Status.Applied.Objects.Console.ImageTag
			template.ImageURL = template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.Console.ImageContext != "" {
			template.ImageContext = cr.Status.Applied.Objects.Console.ImageContext
			template.ImageURL = template.ImageContext + "/" + template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.Console.GitHooks != nil {
			template.GitHooks = *cr.Status.Applied.Objects.Console.GitHooks.DeepCopy()
			if template.GitHooks.MountPath == "" {
				template.GitHooks.MountPath = constants.GitHooksDefaultDir
			}
		}
		// JVM configuration
		if cr.Status.Applied.Objects.Console.Jvm != nil {
			template.Jvm = *cr.Status.Applied.Objects.Console.Jvm.DeepCopy()
		}
		// Simplified mode configuration
		if enabled, err := strconv.ParseBool(getSpecEnv(cr.Status.Applied.Objects.Console.Env, "ORG_APPFORMER_SIMPLIFIED_MONITORING_ENABLED")); err == nil {
			template.Simplified = enabled
		}
	}
	return template
}

func getSpecEnv(envs []corev1.EnvVar, name string) string {
	for _, env := range envs {
		if env.Name == name {
			return env.Value
		}
	}
	return ""
}

func getSmartRouterTemplate(cr *api.KieApp) api.SmartRouterTemplate {
	template := api.SmartRouterTemplate{}
	if cr.Status.Applied.Objects.SmartRouter != nil {
		envConstants, hasEnv := constants.EnvironmentConstants[cr.Status.Applied.Environment]
		if !hasEnv {
			return template
		}
		// Set replicas
		if cr.Status.Applied.Objects.SmartRouter.Replicas == nil {
			cr.Status.Applied.Objects.SmartRouter.Replicas = &envConstants.Replica.SmartRouter.Replicas
		}
		template.Replicas = *cr.Status.Applied.Objects.SmartRouter.Replicas
		if cr.Status.Applied.Objects.SmartRouter.KeystoreSecret == "" {
			template.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, strings.Join([]string{cr.Status.Applied.CommonConfig.ApplicationName, "smartrouter"}, "-"))
		} else {
			template.KeystoreSecret = cr.Status.Applied.Objects.SmartRouter.KeystoreSecret
		}
		if cr.Status.Applied.Objects.SmartRouter.Protocol == "" {
			template.Protocol = constants.SmartRouterProtocol
		} else {
			template.Protocol = cr.Status.Applied.Objects.SmartRouter.Protocol
		}
		template.UseExternalRoute = cr.Status.Applied.Objects.SmartRouter.UseExternalRoute
		template.StorageClassName = cr.Status.Applied.Objects.SmartRouter.StorageClassName
		cMajor, _, _ := GetMajorMinorMicro(cr.Status.Applied.Version)
		template.ImageURL = constants.ImageRegistry + "/" + constants.RhpamPrefix + "-" + cMajor + "/" + constants.RhpamPrefix + "-smartrouter" + constants.RhelVersion + ":" + cr.Status.Applied.Version
		if !cr.Status.Applied.UseImageTags {
			if val, exists := os.LookupEnv(constants.PamSmartRouterVar + cr.Status.Applied.Version); exists {
				template.ImageURL = val
			}
			template.OmitImageStream = true
		}
		template.Image, template.ImageTag, template.ImageContext = GetImage(template.ImageURL)

		if cr.Status.Applied.Objects.SmartRouter.Image != "" {
			template.Image = cr.Status.Applied.Objects.SmartRouter.Image
			template.ImageURL = template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.SmartRouter.ImageTag != "" {
			template.ImageTag = cr.Status.Applied.Objects.SmartRouter.ImageTag
			template.ImageURL = template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.SmartRouter.ImageContext != "" {
			template.ImageContext = cr.Status.Applied.Objects.SmartRouter.ImageContext
			template.ImageURL = template.ImageContext + "/" + template.Image + ":" + template.ImageTag
			template.OmitImageStream = false
		}
	}
	return template
}

// GetImage ...
func GetImage(imageURL string) (image, imageTag, imageContext string) {
	urlParts := strings.Split(imageURL, "/")
	if len(urlParts) > 1 {
		imageContext = urlParts[len(urlParts)-2]
	}
	imageAndTag := urlParts[len(urlParts)-1]
	imageParts := strings.Split(imageAndTag, ":")
	image = imageParts[0]
	if len(imageParts) > 1 {
		imageTag = imageParts[len(imageParts)-1]
	}
	return image, imageTag, imageContext
}

func setReplicas(objectReplicas *int32, replicaConstant api.Replicas, hasEnv bool) (replicas int32, denyScale bool) {
	if objectReplicas != nil {
		if hasEnv && replicaConstant.DenyScale && *objectReplicas != replicaConstant.Replicas {
			log.Warnf("scaling not allowed for this environment, setting to default of %d", replicaConstant.Replicas)
			return replicaConstant.Replicas, true
		}
		return *objectReplicas, false
	}
	if hasEnv {
		return replicaConstant.Replicas, false
	}
	log.Warnf("no replicas settings for this environment, defaulting to %d", replicas)
	return int32(1), denyScale
}

// serverSortBlanks moves blank names to the end
func serverSortBlanks(serverSets []api.KieServerSet) []api.KieServerSet {
	var newSets []api.KieServerSet
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
func getServersConfig(cr *api.KieApp) ([]api.ServerTemplate, error) {
	var servers []api.ServerTemplate
	serverReplicas := Pint32(1)
	envConstants, hasEnv := constants.EnvironmentConstants[cr.Status.Applied.Environment]
	if hasEnv {
		serverReplicas = &envConstants.Replica.Server.Replicas
	}
	product := GetProduct(cr.Status.Applied.Environment)
	usedNames := map[string]bool{}
	serverSlice := cr.Status.Applied.Objects.Servers
	for index := range serverSlice {
		serverSet := &serverSlice[index]
		if serverSet.Deployments == nil {
			serverSet.Deployments = Pint(constants.DefaultKieDeployments)
		}
		for i := 0; i < *serverSet.Deployments; i++ {
			name := getKieDeploymentName(cr.Status.Applied.CommonConfig.ApplicationName, serverSet.Name, 0, i)
			if usedNames[name] {
				return []api.ServerTemplate{}, fmt.Errorf("duplicate kieserver name %s", name)
			}
			usedNames[name] = true
			template := api.ServerTemplate{
				KieName:          name,
				KieServerID:      name,
				Build:            getBuildConfig(product, cr, serverSet),
				KeystoreSecret:   serverSet.KeystoreSecret,
				StorageClassName: serverSet.StorageClassName,
			}

			if cr.Status.Applied.Objects.Console == nil || cr.Status.Applied.Environment == api.RhdmProductionImmutable {
				template.OmitConsole = true
			}
			if serverSet.ID != "" {
				template.KieServerID = serverSet.ID
			}
			if serverSet.Build != nil {
				if *serverSet.Deployments > 1 {
					return []api.ServerTemplate{}, fmt.Errorf("Cannot request %v deployments for a build", *serverSet.Deployments)
				}
				template.From = api.ImageObjRef{
					Kind: "ImageStreamTag",
					ObjectReference: api.ObjectReference{
						Name:      fmt.Sprintf("%s:latest", serverSet.Name),
						Namespace: "",
					},
				}
			} else {
				template.From, template.OmitImageStream, template.ImageURL = getDefaultKieServerImage(product, cr, serverSet, false)
			}

			// Set replicas
			if serverSet.Replicas == nil {
				serverSet.Replicas = serverReplicas
			}
			template.Replicas = *serverSet.Replicas

			// if, SmartRouter object is nil, ignore it
			// get smart router protocol configuration
			if cr.Status.Applied.Objects.SmartRouter != nil {
				if cr.Status.Applied.Objects.SmartRouter.Protocol == "" {
					template.SmartRouter.Protocol = constants.SmartRouterProtocol
				} else {
					template.SmartRouter.Protocol = cr.Status.Applied.Objects.SmartRouter.Protocol
				}
			}

			dbConfig, err := getDatabaseConfig(cr.Status.Applied.Environment, serverSet.Database)
			if err != nil {
				return servers, err
			}
			if dbConfig != nil {
				template.Database = *dbConfig
			}

			jmsConfig, err := getJmsConfig(cr.Status.Applied.Environment, serverSet.Jms)
			if err != nil {
				return servers, err
			}
			if jmsConfig != nil {
				template.Jms = *jmsConfig
			}

			instanceTemplate := template.DeepCopy()
			if instanceTemplate.KeystoreSecret == "" {
				instanceTemplate.KeystoreSecret = fmt.Sprintf(constants.KeystoreSecret, instanceTemplate.KieName)
			}

			// JVM configuration
			if serverSet.Jvm != nil {
				instanceTemplate.Jvm = *serverSet.Jvm.DeepCopy()
			}

			servers = append(servers, *instanceTemplate)
		}

	}
	return servers, nil
}

// GetServerSet retrieves to correct ServerSet for processing and the DeploymentName
func GetServerSet(cr *api.KieApp, requestedIndex int) (serverSet api.KieServerSet, kieName string) {
	count := 0
	unnamedSets := 0
	for _, thisServerSet := range cr.Status.Applied.Objects.Servers {
		for relativeIndex := 0; relativeIndex < *thisServerSet.Deployments; relativeIndex++ {
			if count == requestedIndex {
				serverSet = thisServerSet
				kieName = getKieDeploymentName(cr.Status.Applied.CommonConfig.ApplicationName, serverSet.Name, unnamedSets, relativeIndex)
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

func setKieSetNames(spec *api.KieAppSpec) {
	spec.Objects.Servers = serverSortBlanks(spec.Objects.Servers)
	for index := range spec.Objects.Servers {
		if spec.Objects.Servers[index].Name == "" {
			spec.Objects.Servers[index].Name = getKieSetName(spec, index)
		}
	}
}

func getKieSetName(spec *api.KieAppSpec, index int) string {
	unsetNames := 0
	for i := 0; i < len(spec.Objects.Servers); i++ {
		serverSetName := getKieDeploymentName(spec.CommonConfig.ApplicationName, spec.Objects.Servers[index].Name, unsetNames, 0)
		if !usedServerSetName(spec.Objects.Servers, serverSetName) {
			return serverSetName
		}
		unsetNames++
	}
	return ""
}

// ConsolidateObjects construct all CustomObjects prior to creation
func ConsolidateObjects(env api.Environment, cr *api.KieApp) api.Environment {
	if cr.Status.Applied.Objects.Console != nil {
		env.Console = ConstructObject(env.Console, cr.Status.Applied.Objects.Console.KieAppObject)
	}
	if cr.Status.Applied.Objects.SmartRouter != nil {
		env.SmartRouter = ConstructObject(env.SmartRouter, cr.Status.Applied.Objects.SmartRouter.KieAppObject)
	}
	for index := range env.Servers {
		serverSet, _ := GetServerSet(cr, index)
		env.Servers[index] = ConstructObject(env.Servers[index], serverSet.KieAppObject)
		// apply the build config envs provided through cr to the given kieSerever BuildConfig Spec envs
		for bcindex := range env.Servers[index].BuildConfigs {
			env.Servers[index].BuildConfigs[bcindex].Spec.Strategy.SourceStrategy.Env = shared.EnvOverride(
				env.Servers[index].BuildConfigs[bcindex].Spec.Strategy.SourceStrategy.Env,
				serverSet.Build.Env)
		}
	}
	return env
}

// ConstructObject returns an object after merging the environment object and the one defined in the CR
func ConstructObject(object api.CustomObject, appObject api.KieAppObject) api.CustomObject {
	for dcIndex, dc := range object.DeploymentConfigs {
		for containerIndex, c := range dc.Spec.Template.Spec.Containers {
			c.Env = shared.EnvOverride(c.Env, appObject.Env)
			if appObject.Resources != nil {
				err := mergo.Merge(&c.Resources, *appObject.Resources, mergo.WithOverride)
				if err != nil {
					log.Error("Error merging interfaces. ", err)
				}
			}
			dc.Spec.Template.Spec.Containers[containerIndex] = c
		}
		object.DeploymentConfigs[dcIndex] = dc
	}
	return object
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

func getBuildConfig(product string, cr *api.KieApp, serverSet *api.KieServerSet) api.BuildTemplate {
	if serverSet.Build == nil {
		return api.BuildTemplate{}
	}
	buildTemplate := api.BuildTemplate{}

	if serverSet.Build.ExtensionImageStreamTag != "" {
		if serverSet.Build.ExtensionImageStreamTagNamespace == "" {
			serverSet.Build.ExtensionImageStreamTagNamespace = constants.ImageStreamNamespace
			log.Debugf("Extension Image Stream Tag set but no namespace set, defaulting to %s", serverSet.Build.ExtensionImageStreamTagNamespace)
		}
		if serverSet.Build.ExtensionImageInstallDir == "" {
			serverSet.Build.ExtensionImageInstallDir = constants.DefaultExtensionImageInstallDir
		} else {
			log.Debugf("Extension Image Install Dir set to %s, be cautions when updating this parameter.", serverSet.Build.ExtensionImageInstallDir)
		}
		// JDBC extension image build template
		buildTemplate = api.BuildTemplate{
			ExtensionImageStreamTag:          serverSet.Build.ExtensionImageStreamTag,
			ExtensionImageStreamTagNamespace: serverSet.Build.ExtensionImageStreamTagNamespace,
			ExtensionImageInstallDir:         serverSet.Build.ExtensionImageInstallDir,
		}

	} else {
		// build app from source template
		buildTemplate = api.BuildTemplate{
			GitSource:                    serverSet.Build.GitSource,
			GitHubWebhookSecret:          getWebhookSecret(api.GitHubWebhook, serverSet.Build.Webhooks),
			GenericWebhookSecret:         getWebhookSecret(api.GenericWebhook, serverSet.Build.Webhooks),
			KieServerContainerDeployment: serverSet.Build.KieServerContainerDeployment,
			MavenMirrorURL:               serverSet.Build.MavenMirrorURL,
			ArtifactDir:                  serverSet.Build.ArtifactDir,
		}
	}

	buildTemplate.From, _, _ = getDefaultKieServerImage(product, cr, serverSet, true)
	if serverSet.Build.From != nil {
		buildTemplate.From = *serverSet.Build.From
	}

	return buildTemplate
}

func getDefaultKieServerImage(product string, cr *api.KieApp, serverSet *api.KieServerSet, forBuild bool) (from api.ImageObjRef, omitImageTrigger bool, imageURL string) {
	if serverSet.From != nil {
		return *serverSet.From, omitImageTrigger, imageURL
	}
	envVar := constants.PamKieImageVar + cr.Status.Applied.Version
	if product == constants.RhdmPrefix {
		envVar = constants.DmKieImageVar + cr.Status.Applied.Version
	}

	cMajor, _, _ := GetMajorMinorMicro(cr.Status.Applied.Version)
	imageURL = constants.ImageRegistry + "/" + product + "-" + cMajor + "/" + product + "-kieserver" + constants.RhelVersion + ":" + cr.Status.Applied.Version
	if !cr.Status.Applied.UseImageTags && !forBuild {
		if val, exists := os.LookupEnv(envVar); exists {
			imageURL = val
		}
		omitImageTrigger = true
	}
	image, imageTag, imageContext := GetImage(imageURL)

	if serverSet.Image != "" {
		image = serverSet.Image
		imageURL = image + ":" + imageTag
		omitImageTrigger = false
	}
	if serverSet.ImageTag != "" {
		imageTag = serverSet.ImageTag
		imageURL = image + ":" + imageTag
		omitImageTrigger = false
	}
	if serverSet.ImageContext != "" {
		imageContext = serverSet.ImageContext
		imageURL = imageContext + "/" + image + ":" + imageTag
		omitImageTrigger = false
	}

	return api.ImageObjRef{
		Kind: "ImageStreamTag",
		ObjectReference: api.ObjectReference{
			Name:      image + ":" + imageTag,
			Namespace: constants.ImageStreamNamespace,
		},
	}, omitImageTrigger, imageURL
}

func getDatabaseConfig(environment api.EnvironmentType, database *api.DatabaseObject) (*api.DatabaseObject, error) {
	envConstants := constants.EnvironmentConstants[environment]
	if envConstants == nil {
		return nil, nil
	}
	defaultDB := envConstants.Database
	if database == nil {
		return defaultDB, nil
	}

	if database.Type == api.DatabaseExternal && database.ExternalConfig == nil {
		return nil, fmt.Errorf("external database configuration is mandatory for external database type")
	}

	if database.Size == "" && defaultDB != nil {
		resultDB := *database.DeepCopy()
		resultDB.Size = defaultDB.Size
		return &resultDB, nil
	}
	return database, nil
}

func getJmsConfig(environment api.EnvironmentType, jms *api.KieAppJmsObject) (*api.KieAppJmsObject, error) {
	envConstants := constants.EnvironmentConstants[environment]
	if envConstants == nil || jms == nil || !jms.EnableIntegration {
		return nil, nil
	}
	if jms.AMQSecretName != "" && jms.AMQKeystoreName != "" && jms.AMQKeystorePassword != "" &&
		jms.AMQTruststoreName != "" && jms.AMQTruststorePassword != "" {
		jms.AMQEnableSSL = true
	}
	t := true
	if jms.Executor == nil {
		jms.Executor = &t
	}
	if jms.AuditTransacted == nil {
		jms.AuditTransacted = &t
	}

	// if enabled, prepare the default values
	defaultJms := api.KieAppJmsObject{
		QueueExecutor: "queue/KIE.SERVER.EXECUTOR",
		QueueRequest:  "queue/KIE.SERVER.REQUEST",
		QueueResponse: "queue/KIE.SERVER.RESPONSE",
		QueueSignal:   "queue/KIE.SERVER.SIGNAL",
		QueueAudit:    "queue/KIE.SERVER.AUDIT",
		Username:      "user" + string(shared.GeneratePassword(4)),
		Password:      string(shared.GeneratePassword(8)),
	}

	queuesList := []string{
		getDefaultQueue(*jms.Executor, defaultJms.QueueExecutor, jms.QueueExecutor),
		getDefaultQueue(true, defaultJms.QueueRequest, jms.QueueRequest),
		getDefaultQueue(true, defaultJms.QueueResponse, jms.QueueResponse),
		getDefaultQueue(jms.EnableSignal, defaultJms.QueueSignal, jms.QueueSignal),
		getDefaultQueue(jms.EnableAudit, defaultJms.QueueAudit, jms.QueueAudit),
	}

	// clean empty values
	for i, queue := range queuesList {
		if queue == "" {
			queuesList = append(queuesList[:i], queuesList[i+1:]...)
			break
		}
	}
	defaultJms.AMQQueues = strings.Join(queuesList[:], ", ")

	// merge the defaultJms into jms, preserving the values set by cr
	if err := mergo.Merge(jms, defaultJms); err != nil {
		return jms, err
	}
	return jms, nil
}

func getDefaultQueue(append bool, defaultJmsQueue string, jmsQueue string) string {
	if append {
		if jmsQueue == "" {
			return defaultJmsQueue
		}
		return jmsQueue
	}
	return ""
}

func setPasswords(spec *api.KieAppSpec, isTrialEnv bool) {
	passwords := []*string{
		&spec.CommonConfig.KeyStorePassword,
		&spec.CommonConfig.AdminPassword,
		&spec.CommonConfig.DBPassword,
		&spec.CommonConfig.AMQPassword,
		&spec.CommonConfig.AMQClusterPassword,
	}
	for i := range passwords {
		if len(*passwords[i]) > 0 {
			continue
		}
		if isTrialEnv {
			*passwords[i] = constants.DefaultPassword
		} else {
			*passwords[i] = string(shared.GeneratePassword(8))
		}
	}
}

func getWebhookSecret(webhookType api.WebhookType, webhooks []api.WebhookSecret) string {
	for _, webhook := range webhooks {
		if webhook.Type == webhookType {
			return webhook.Secret
		}
	}
	return ""
}

// important to parse template first with this function, before unmarshalling into object
func loadYaml(service kubernetes.PlatformService, filename, productVersion, namespace string, env api.EnvTemplate) ([]byte, error) {
	// prepend specified product version dir to filepath
	filename = strings.Join([]string{productVersion, filename}, "/")
	if _, _, useEmbedded := UseEmbeddedFiles(service); useEmbedded {
		box := packr.New("config", "../../../../config")
		if !box.HasDir(productVersion) {
			return nil, fmt.Errorf("Product version %s configs are not available in this Operator, %s", productVersion, version.Version)
		}
		if box.Has(filename) {
			yamlString, err := box.FindString(filename)
			if err != nil {
				return nil, err
			}
			return parseTemplate(env, yamlString)
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
	return parseTemplate(env, configMap.Data[file])
}

func parseTemplate(env api.EnvTemplate, objYaml string) ([]byte, error) {
	var b bytes.Buffer

	tmpl, err := template.New(env.ApplicationName).Delims("[[", "]]").Parse(objYaml)
	if err != nil {
		log.Error("Error creating new Go template.")
		return []byte{}, err
	}

	// template replacement
	err = tmpl.Execute(&b, env)
	if err != nil {
		log.Error("Error applying Go template.")
		return []byte{}, err
	}

	return b.Bytes(), nil
}

// convertToConfigMapName ...
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

// getCMListfromBox reads the files under the config folder ...
func getCMListfromBox(box *packr.Box) map[string][]map[string]string {
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
	return cmList
}

// ConfigMapsFromFile reads the files under the config folder and creates
// configmaps in the given namespace. It sets OwnerRef to operator deployment.
func ConfigMapsFromFile(myDep *appsv1.Deployment, ns string, scheme *runtime.Scheme) (configMaps []corev1.ConfigMap) {
	box := packr.New("config", "../../../../config")
	cmList := getCMListfromBox(box)
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
					api.SchemeGroupVersion.Group: version.Version,
				},
			},
			Data: cmData,
		}

		cm.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		cm.SetOwnerReferences(myDep.GetOwnerReferences())
		configMaps = append(configMaps, cm)
	}
	return configMaps
}

// UseEmbeddedFiles checks environment variables WATCH_NAMESPACE & OPERATOR_NAME
func UseEmbeddedFiles(service kubernetes.PlatformService) (opName string, depNameSpace string, useEmbedded bool) {
	namespace := os.Getenv(constants.NameSpaceEnv)
	name := os.Getenv(constants.OpNameEnv)
	if service.IsMockService() || namespace == "" || name == "" {
		return name, namespace, true
	}
	return name, namespace, false
}

// Pint returns a pointer to an integer
func Pint(i int) *int {
	return &i
}

// Pint32 returns a pointer to an integer
func Pint32(i int32) *int32 {
	return &i
}

// Pbool returns a pointer to a boolean
func Pbool(b bool) *bool {
	return &b
}

// GetProduct ...
func GetProduct(env api.EnvironmentType) (product string) {
	envConstants := constants.EnvironmentConstants[env]
	if envConstants != nil {
		product = envConstants.App.Product
	}
	return
}

func usedServerSetName(servers []api.KieServerSet, serverSetName string) bool {
	for _, server := range servers {
		if server.Name == serverSetName {
			return true
		}
	}
	return false
}

// SetDefaults set default values where not provided
func SetDefaults(cr *api.KieApp) {
	if cr.GetAnnotations() == nil {
		cr.SetAnnotations(map[string]string{
			api.SchemeGroupVersion.Group: version.Version,
		})
	}

	// retain certain items from status... e.g. version, usernames, passwords, etc
	// everything else in status should be recreated with each reconcile.
	specApply := cr.Spec.DeepCopy()

	if !isImmutable(cr) && specApply.Objects.Console == nil {
		specApply.Objects.Console = &api.ConsoleObject{
			KieAppObject: api.KieAppObject{},
		}
	}

	if len(specApply.Version) == 0 {
		specApply.Version = constants.CurrentVersion
		if len(cr.Status.Applied.Version) != 0 {
			specApply.Version = cr.Status.Applied.Version
		}
	}
	if err := mergo.Merge(&specApply.CommonConfig, cr.Status.Applied.CommonConfig); err != nil {
		log.Error(err)
	}
	if len(specApply.CommonConfig.ApplicationName) == 0 {
		specApply.CommonConfig.ApplicationName = cr.Name
	}
	if len(specApply.CommonConfig.AdminUser) == 0 {
		specApply.CommonConfig.AdminUser = constants.DefaultAdminUser
	}
	if len(specApply.Objects.Servers) == 0 {
		specApply.Objects.Servers = []api.KieServerSet{{Deployments: Pint(constants.DefaultKieDeployments)}}
	}
	setKieSetNames(specApply)

	for index := range specApply.Objects.Servers {
		addWebhookTypes(specApply.Objects.Servers[index].Build)
		for _, statusServer := range cr.Status.Applied.Objects.Servers {
			retainAppliedPwds(&specApply.Objects.Servers[index], statusServer)
		}
		addWebhookPwds(specApply.Objects.Servers[index].Build)
		checkJvmOnServer(&specApply.Objects.Servers[index])
		setResourcesDefault(&specApply.Objects.Servers[index].KieAppObject, constants.ServersLimits, constants.ServerRequests)
	}

	if specApply.Objects.Console != nil {
		checkJvmOnConsole(specApply.Objects.Console)
		if strings.Contains(string(specApply.Environment), "authoring") {
			setResourcesDefault(&specApply.Objects.Console.KieAppObject, constants.ConsoleAuthoringLimits, constants.ConsoleAuthoringRequests)
		} else if strings.Contains(string(specApply.Environment), "production") {
			setResourcesDefault(&specApply.Objects.Console.KieAppObject, constants.ConsoleProdLimits, constants.ConsoleProdRequests)
		}
	}

	if specApply.Objects.SmartRouter != nil {
		setResourcesDefault(&specApply.Objects.SmartRouter.KieAppObject, constants.SmartRouterLimits, constants.SmartRouterRequests)
	}

	isTrialEnv := strings.HasSuffix(string(specApply.Environment), constants.TrialEnvSuffix)
	setPasswords(specApply, isTrialEnv)

	cr.Status.Applied = *specApply
}

func checkJvmOnConsole(console *api.ConsoleObject) {
	if console.Jvm == nil {
		console.Jvm = &api.JvmObject{}
	}
	setJvmDefault(console.Jvm)
}

func setJvmDefault(jvm *api.JvmObject) {
	if jvm != nil {
		if jvm.JavaMaxMemRatio == nil {
			jvm.JavaMaxMemRatio = Pint32(80)
		}
		if jvm.JavaInitialMemRatio == nil {
			jvm.JavaInitialMemRatio = Pint32(25)
		}
	}
}

func checkJvmOnServer(server *api.KieServerSet) {
	if server.Jvm == nil {
		server.Jvm = &api.JvmObject{}
	}
	setJvmDefault(server.Jvm)
}

func setResourcesDefault(kieObject *api.KieAppObject, limits, requests map[string]string) {
	if kieObject.Resources == nil {
		kieObject.Resources = &corev1.ResourceRequirements{}
	}
	if kieObject.Resources.Limits == nil {
		kieObject.Resources.Limits = corev1.ResourceList{corev1.ResourceCPU: createResourceQuantity(limits["CPU"]), corev1.ResourceMemory: createResourceQuantity(limits["MEM"])}
	}
	if kieObject.Resources.Requests == nil {
		kieObject.Resources.Requests = corev1.ResourceList{corev1.ResourceCPU: createResourceQuantity(requests["CPU"]), corev1.ResourceMemory: createResourceQuantity(requests["MEM"])}
	}
	if kieObject.Resources.Limits.Cpu() == nil {
		kieObject.Resources.Limits.Cpu().Add(createResourceQuantity(limits["CPU"]))
	}
	if kieObject.Resources.Limits.Memory() == nil {
		kieObject.Resources.Limits.Memory().Add(createResourceQuantity(limits["MEM"]))
	}
	if kieObject.Resources.Requests.Cpu() == nil {
		kieObject.Resources.Requests.Cpu().Add(createResourceQuantity(requests["CPU"]))
	}
	if kieObject.Resources.Requests.Memory() == nil {
		kieObject.Resources.Requests.Memory().Add(createResourceQuantity(requests["MEM"]))
	}
}

func createResourceQuantity(constraint string) resource.Quantity {
	item, err := resource.ParseQuantity(constraint)
	if err != nil {
		log.Error(err)
	}
	return item
}

func addWebhookTypes(buildObject *api.KieAppBuildObject) {
	if buildObject == nil {
		return
	}
	whTypes := []api.WebhookType{api.GenericWebhook, api.GitHubWebhook}
	for _, whType := range whTypes {
		missing := true
		for _, whSecret := range buildObject.Webhooks {
			if whSecret.Type == whType {
				missing = false
			}
		}
		if missing {
			buildObject.Webhooks = append(buildObject.Webhooks, api.WebhookSecret{Type: whType})
		}
	}
}

func retainAppliedPwds(dst *api.KieServerSet, src api.KieServerSet) {
	if dst.Name == src.Name {
		retainJMSPwds(dst.Jms, src.Jms)
		retainWebhookSecrets(dst.Build, src.Build)
	}
}

func retainJMSPwds(dstJms *api.KieAppJmsObject, srcJms *api.KieAppJmsObject) {
	if dstJms == nil || srcJms == nil {
		return
	}
	if dstJms.Username == "" {
		dstJms.Username = srcJms.Username
	}
	if dstJms.Password == "" {
		dstJms.Password = srcJms.Password
	}
}

func retainWebhookSecrets(dstBuild *api.KieAppBuildObject, srcBuild *api.KieAppBuildObject) {
	if dstBuild == nil || srcBuild == nil {
		return
	}
	for whIndex := range dstBuild.Webhooks {
		for _, srcWh := range srcBuild.Webhooks {
			if dstBuild.Webhooks[whIndex].Type == srcWh.Type &&
				dstBuild.Webhooks[whIndex].Secret == "" {
				dstBuild.Webhooks[whIndex].Secret = srcWh.Secret
			}
		}
	}
}

func addWebhookPwds(buildObject *api.KieAppBuildObject) {
	if buildObject == nil {
		return
	}
	for whIndex := range buildObject.Webhooks {
		if buildObject.Webhooks[whIndex].Secret == "" {
			buildObject.Webhooks[whIndex].Secret = string(shared.GeneratePassword(8))
		}
	}
}

func getProcessMigrationTemplate(cr *api.KieApp, serversConfig []api.ServerTemplate) (processMigrationTemplate *api.ProcessMigrationTemplate, err error) {
	if deployProcessMigration(cr) {
		processMigrationTemplate = &api.ProcessMigrationTemplate{}
		processMigrationTemplate.ImageURL = constants.ProcessMigrationDefaultImageURL + ":" + cr.Status.Applied.Version
		if val, exists := os.LookupEnv(constants.PamProcessMigrationVar + cr.Status.Applied.Version); exists && !cr.Status.Applied.UseImageTags {
			processMigrationTemplate.ImageURL = val
			processMigrationTemplate.OmitImageStream = true
		}
		processMigrationTemplate.Image, processMigrationTemplate.ImageTag, processMigrationTemplate.ImageContext = GetImage(processMigrationTemplate.ImageURL)
		if cr.Status.Applied.Objects.ProcessMigration.Image != "" {
			processMigrationTemplate.Image = cr.Status.Applied.Objects.ProcessMigration.Image
			processMigrationTemplate.ImageURL = processMigrationTemplate.Image + ":" + processMigrationTemplate.ImageTag
			processMigrationTemplate.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.ProcessMigration.ImageTag != "" {
			processMigrationTemplate.ImageTag = cr.Status.Applied.Objects.ProcessMigration.ImageTag
			processMigrationTemplate.ImageURL = processMigrationTemplate.Image + ":" + processMigrationTemplate.ImageTag
			processMigrationTemplate.OmitImageStream = false
		}
		if cr.Status.Applied.Objects.ProcessMigration.ImageContext != "" {
			processMigrationTemplate.ImageContext = cr.Status.Applied.Objects.ProcessMigration.ImageContext
			processMigrationTemplate.ImageURL = processMigrationTemplate.ImageContext + "/" + processMigrationTemplate.Image + ":" + processMigrationTemplate.ImageTag
			processMigrationTemplate.OmitImageStream = false
		}
		for _, sc := range serversConfig {
			processMigrationTemplate.KieServerClients = append(processMigrationTemplate.KieServerClients, api.KieServerClient{
				Host:     fmt.Sprintf("http://%s:8080/services/rest/server", sc.KieName),
				Username: cr.Status.Applied.CommonConfig.AdminUser,
				Password: cr.Status.Applied.CommonConfig.AdminPassword,
			})
		}
		if cr.Status.Applied.Objects.ProcessMigration.Database.Type == "" {
			processMigrationTemplate.Database.Type = constants.DefaultProcessMigrationDatabaseType
		} else if cr.Status.Applied.Objects.ProcessMigration.Database.Type == api.DatabaseExternal &&
			cr.Status.Applied.Objects.ProcessMigration.Database.ExternalConfig == nil {
			return nil, fmt.Errorf("external database configuration is mandatory for external database type of process migration")
		} else {
			processMigrationTemplate.Database = *cr.Status.Applied.Objects.ProcessMigration.Database.DeepCopy()
		}
	}
	return processMigrationTemplate, nil
}

func mergeProcessMigration(service kubernetes.PlatformService, cr *api.KieApp, env api.Environment, envTemplate api.EnvTemplate) (api.Environment, error) {
	var processMigrationEnv api.Environment
	if deployProcessMigration(cr) {
		processMigrationEnv, err := loadProcessMigrationFromFile("pim/process-migration.yaml", service, cr, envTemplate)
		if err != nil {
			return api.Environment{}, err
		}
		env.ProcessMigration = mergeCustomObject(env.ProcessMigration, processMigrationEnv.ProcessMigration)
		env, err = mergeProcessMigrationDB(service, cr, env, envTemplate)
		if err != nil {
			return api.Environment{}, nil
		}
		if api.RhpamTrial == cr.Spec.Environment {
			pimEnv, err := loadProcessMigrationFromFile("pim/process-migration-trial.yaml", service, cr, envTemplate)
			if err != nil {
				return api.Environment{}, nil
			}
			env.ProcessMigration = mergeCustomObject(env.ProcessMigration, pimEnv.ProcessMigration)
		}
	} else {
		processMigrationEnv.ProcessMigration.Omit = true
	}

	return env, nil
}

func loadProcessMigrationFromFile(filename string, service kubernetes.PlatformService, cr *api.KieApp, envTemplate api.EnvTemplate) (api.Environment, error) {
	var pimEnv api.Environment
	yamlBytes, err := loadYaml(service, filename, cr.Status.Applied.Version, cr.Namespace, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	err = yaml.Unmarshal(yamlBytes, &pimEnv)
	if err != nil {
		return api.Environment{}, err
	}
	return pimEnv, nil
}

func mergeProcessMigrationDB(service kubernetes.PlatformService, cr *api.KieApp, env api.Environment, envTemplate api.EnvTemplate) (api.Environment, error) {
	if envTemplate.ProcessMigration.Database.Type == api.DatabaseH2 {
		return env, nil
	}
	yamlBytes, err := loadYaml(service, fmt.Sprintf("dbs/pim/%s.yaml", envTemplate.ProcessMigration.Database.Type), cr.Status.Applied.Version, cr.Namespace, envTemplate)
	if err != nil {
		return api.Environment{}, err
	}
	var dbEnv api.Environment
	err = yaml.Unmarshal(yamlBytes, &dbEnv)
	if err != nil {
		return api.Environment{}, err
	}
	env.ProcessMigration = mergeCustomObject(env.ProcessMigration, dbEnv.ProcessMigration)

	return env, nil
}

func isRHPAM(cr *api.KieApp) bool {
	switch cr.Status.Applied.Environment {
	case api.RhpamTrial, api.RhpamAuthoring, api.RhpamAuthoringHA, api.RhpamProduction, api.RhpamProductionImmutable:
		return true
	}
	return false
}

func isImmutable(cr *api.KieApp) bool {
	switch cr.Status.Applied.Environment {
	case api.RhdmProductionImmutable, api.RhpamProductionImmutable:
		return true
	}
	return false
}

func deployProcessMigration(cr *api.KieApp) bool {
	return isGE78(cr) && cr.Status.Applied.Objects.ProcessMigration != nil && isRHPAM(cr)
}

func isGE78(cr *api.KieApp) bool {
	return semver.Compare(semver.MajorMinor("v"+cr.Status.Applied.Version), "v7.8") >= 0
}

func getDatabaseDeploymentTemplate(cr *api.KieApp, serversConfig []api.ServerTemplate,
	processMigrationTemplate *api.ProcessMigrationTemplate) []api.DatabaseTemplate {
	var databaseDeploymentTemplate []api.DatabaseTemplate
	if serversConfig != nil {
		for _, sc := range serversConfig {
			if isDeployDB(sc.Database.Type) {
				databaseDeploymentTemplate = append(databaseDeploymentTemplate, api.DatabaseTemplate{
					InternalDatabaseObject: sc.Database.InternalDatabaseObject,
					ServerName:             sc.KieName,
					Username:               constants.DefaultKieServerDatabaseUsername,
					DatabaseName:           constants.DefaultKieServerDatabaseName,
				})
			}
		}
	}
	if processMigrationTemplate != nil && isDeployDB(processMigrationTemplate.Database.Type) {
		databaseDeploymentTemplate = append(databaseDeploymentTemplate, api.DatabaseTemplate{
			InternalDatabaseObject: processMigrationTemplate.Database.InternalDatabaseObject,
			ServerName:             cr.Name + "-process-migration",
			Username:               constants.DefaultProcessMigrationDatabaseUsername,
			DatabaseName:           constants.DefaultProcessMigrationDatabaseName,
		})
	}
	return databaseDeploymentTemplate
}

func mergeDBDeployment(service kubernetes.PlatformService, cr *api.KieApp, env api.Environment, envTemplate api.EnvTemplate) (api.Environment, error) {
	env.Databases = make([]api.CustomObject, len(envTemplate.Databases))
	dbEnvs := make(map[api.DatabaseType]api.Environment)
	for i, dbTemplate := range envTemplate.Databases {
		if err := loadDBYamls(service, cr, envTemplate, "dbs/%s.yaml", dbTemplate.Type, dbEnvs); err != nil {
			return api.Environment{}, err
		}
		deploymentName := dbTemplate.ServerName + "-" + string(dbTemplate.Type)
		for _, db := range dbEnvs[dbTemplate.Type].Databases {
			if len(db.DeploymentConfigs) == 0 {
				continue
			}
			if deploymentName == db.DeploymentConfigs[0].ObjectMeta.Name {
				env.Databases[i] = mergeCustomObject(env.Databases[i], db)
			}
		}
	}
	return env, nil
}

func loadDBYamls(service kubernetes.PlatformService, cr *api.KieApp, envTemplate api.EnvTemplate,
	dbTemplates string, dbType api.DatabaseType, dbEnvs map[api.DatabaseType]api.Environment) error {
	if _, loadedDB := dbEnvs[dbType]; !loadedDB {
		yamlBytes, err := loadYaml(service, fmt.Sprintf(dbTemplates, dbType), cr.Status.Applied.Version, cr.Namespace, envTemplate)
		if err != nil {
			return err
		}
		var dbEnv api.Environment
		err = yaml.Unmarshal(yamlBytes, &dbEnv)
		if err != nil {
			return err
		}
		dbEnvs[dbType] = dbEnv
	}
	return nil
}

func isDeployDB(dbType api.DatabaseType) bool {
	switch dbType {
	case api.DatabaseMySQL, api.DatabasePostgreSQL:
		return true
	}
	return false
}
