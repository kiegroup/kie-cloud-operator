package kieapp

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/status"
	"github.com/kiegroup/kie-cloud-operator/version"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logs.GetLogger("kieapp.controller")

// Reconciler reconciles a KieApp object
type Reconciler struct {
	Service v1.PlatformService
}

// Reconcile reads that state of the cluster for a KieApp object and makes changes based on the state read
// and what is in the KieApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (reconciler *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if opName, depNameSpace, useEmbedded := defaults.UseEmbeddedFiles(reconciler.Service); !useEmbedded {
		myDep := &appsv1.Deployment{}
		err := reconciler.Service.Get(context.TODO(), types.NamespacedName{Namespace: depNameSpace, Name: opName}, myDep)
		if err == nil {
			// Reconcile ConfigMaps
			reconciler.CreateConfigMaps(myDep)
		} else {
			log.Error("Can't properly create ConfigMaps. ", err)
		}
	}

	// Fetch the KieApp instance
	instance := &v1.KieApp{}
	err := reconciler.Service.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reconciler.setFailedStatus(instance, v1.UnknownReason, err)
		return reconcile.Result{}, err
	}
	if instance.GetAnnotations() == nil {
		instance.SetAnnotations(map[string]string{
			v1.SchemeGroupVersion.Group: version.Version,
		})
	}

	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)

	env, rResult, err := reconciler.newEnv(instance)
	if err != nil {
		reconciler.setFailedStatus(instance, v1.ConfigurationErrorReason, err)
		return rResult, err
	}
	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &oappsv1.DeploymentConfigList{}
	err = reconciler.Service.List(context.TODO(), listOps, dcList)
	if err != nil {
		log.Warn("Failed to list dc's. ", err)
		reconciler.setFailedStatus(instance, v1.UnknownReason, err)
		return reconcile.Result{}, err
	}
	dcNames := getDcNames(dcList.Items, instance)
	instance.Status.Deployments = dcNames

	// Update DeploymentConfigs if needed
	var dcUpdates []oappsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		for _, cDc := range env.Console.DeploymentConfigs {
			if dc.Name == cDc.Name {
				dcUpdates = reconciler.dcUpdateCheck(dc, cDc, dcUpdates, instance)
			}
		}
		for _, server := range env.Servers {
			for _, sDc := range server.DeploymentConfigs {
				if dc.Name == sDc.Name {
					dcUpdates = reconciler.dcUpdateCheck(dc, sDc, dcUpdates, instance)
				}
			}
		}
		for _, srDc := range env.Smartrouter.DeploymentConfigs {
			if dc.Name == srDc.Name {
				dcUpdates = reconciler.dcUpdateCheck(dc, srDc, dcUpdates, instance)
			}
		}
		for _, other := range env.Others {
			for _, oDc := range other.DeploymentConfigs {
				if dc.Name == oDc.Name {
					dcUpdates = reconciler.dcUpdateCheck(dc, oDc, dcUpdates, instance)
				}
			}
		}
	}
	if len(dcUpdates) > 0 {
		for _, uDc := range dcUpdates {
			rResult, err := reconciler.UpdateObj(&uDc)
			if err != nil {
				reconciler.setFailedStatus(instance, v1.DeploymentFailedReason, err)
				return rResult, err
			}
		}
		if status.SetProvisioning(instance) {
			return reconciler.UpdateObj(instance)
		}
	}

	// Fetch the cached KieApp instance
	cachedInstance := &v1.KieApp{}
	err = reconciler.Service.GetCached(context.TODO(), request.NamespacedName, cachedInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reconciler.setFailedStatus(instance, v1.UnknownReason, err)
		return reconcile.Result{}, err
	}

	// Update CR if needed
	if !reflect.DeepEqual(instance.Status, cachedInstance.Status) || !reflect.DeepEqual(instance.Spec, cachedInstance.Spec) {
		if status.SetProvisioning(instance) && instance.ResourceVersion == cachedInstance.ResourceVersion {
			return reconciler.UpdateObj(instance)
		}
		return reconcile.Result{Requeue: true}, nil
	}
	if status.SetDeployed(instance) {
		if instance.ResourceVersion == cachedInstance.ResourceVersion {
			return reconciler.UpdateObj(instance)
		}
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (reconciler *Reconciler) setFailedStatus(instance *v1.KieApp, reason v1.ReasonType, err error) {
	status.SetFailed(instance, reason, err)
	_, updateError := reconciler.UpdateObj(instance)
	if updateError != nil {
		log.Warn("Unable to update object after receiving failed status. ", err)
	}
}

// Check ImageStream
func (reconciler *Reconciler) checkImageStreamTag(name, namespace string) bool {
	log := log.With("kind", "ImageStreamTag", "name", name, "namespace", namespace)
	result := strings.Split(name, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	_, err := reconciler.Service.ImageStreamTags(namespace).Get(tagName, metav1.GetOptions{})
	if err != nil {
		log.Debug("Object does not exist")
		return false
	}
	return true
}

// Create local ImageStreamTag
func (reconciler *Reconciler) createLocalImageTag(tagRefName string, cr *v1.KieApp) error {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	version := []byte(cr.Spec.CommonConfig.Version)
	imageName := tagName
	regContext := fmt.Sprintf("%s-%s", cr.Spec.CommonConfig.Product, string(version[0]))

	registryAddress := cr.Spec.ImageRegistry.Registry
	if strings.Contains(result[0], "indexing-openshift") {
		regContext = fmt.Sprintf("%s-7-tech-preview", cr.Spec.CommonConfig.Product)
	} else if strings.Contains(result[0], "amq-broker-7") {
		registryAddress = constants.ImageRegistry
		regContext = "amq-broker-7"
	} else if result[0] == "postgresql" || result[0] == "mysql" {
		registryAddress = constants.ImageRegistry
		regContext = "rhscl"
		pattern := regexp.MustCompile("[0-9]+")
		imageName = fmt.Sprintf("%s-%s-rhel7:%s", result[0], strings.Join(pattern.FindAllString(result[1], -1), ""), "latest")
	}
	registryURL := fmt.Sprintf("%s/%s/%s", registryAddress, regContext, imageName)

	isnew := &oimagev1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tagName,
			Namespace: cr.Namespace,
		},
		Tag: &oimagev1.TagReference{
			Name: result[1],
			From: &corev1.ObjectReference{
				Kind: "DockerImage",
				Name: registryURL,
			},
			ReferencePolicy: oimagev1.TagReferencePolicy{
				Type: oimagev1.LocalTagReferencePolicy,
			},
		},
	}
	isnew.SetGroupVersionKind(oimagev1.SchemeGroupVersion.WithKind("ImageStreamTag"))
	if cr.Spec.ImageRegistry.Insecure {
		isnew.Tag.ImportPolicy = oimagev1.TagImportPolicy{
			Insecure: true,
		}
	}
	log := log.With("kind", isnew.GetObjectKind().GroupVersionKind().Kind, "name", isnew.Name, "from", isnew.Tag.From.Name, "namespace", isnew.Namespace)
	log.Info("Creating")
	_, err := reconciler.Service.ImageStreamTags(isnew.Namespace).Create(isnew)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error("Issue creating object. ", err)
		return err
	}
	return nil
}

func (reconciler *Reconciler) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1.KieApp) []oappsv1.DeploymentConfig {
	log := log.With("kind", current.GetObjectKind().GroupVersionKind().Kind, "name", current.Name, "namespace", current.Namespace)
	update := false
	cContainer := current.Spec.Template.Spec.Containers[0]
	nContainer := new.Spec.Template.Spec.Containers[0]

	if !shared.EnvVarCheck(cContainer.Env, nContainer.Env) {
		log.Debug("Changes detected in 'Env' config.", " OLD - ", cContainer.Env, " NEW - ", nContainer.Env)
		update = true
	}
	if !reflect.DeepEqual(cContainer.Resources, nContainer.Resources) {
		log.Debug("Changes detected in 'Resource' config.", " OLD - ", cContainer.Resources, " NEW - ", nContainer.Resources)
		update = true
	}

	if update {
		dcnew := new
		err := controllerutil.SetControllerReference(cr, &dcnew, reconciler.Service.GetScheme())
		if err != nil {
			log.Error("Error setting controller reference for dc. ", err)
		}
		dcnew.SetNamespace(current.Namespace)
		dcnew.SetResourceVersion(current.ResourceVersion)
		dcnew.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))

		dcUpdates = append(dcUpdates, dcnew)
	}
	return dcUpdates
}

// NewEnv creates an Environment generated from the given KieApp
func (reconciler *Reconciler) newEnv(cr *v1.KieApp) (v1.Environment, reconcile.Result, error) {
	env, err := defaults.GetEnvironment(cr, reconciler.Service)
	if err != nil {
		return env, reconcile.Result{Requeue: true}, err
	}

	// console keystore generation
	if !env.Console.Omit {
		consoleCN := ""
		for _, rt := range env.Console.Routes {
			if checkTLS(rt.Spec.TLS) {
				// use host of first tls route in env template
				consoleCN = reconciler.GetRouteHost(rt, cr)
				cr.Status.ConsoleHost = fmt.Sprintf("https://%s", consoleCN)
				break
			}
		}
		if consoleCN == "" {
			consoleCN = cr.Name
			cr.Status.ConsoleHost = fmt.Sprintf("http://%s", consoleCN)
		}

		defaults.ConfigureHostname(&env.Console, cr, consoleCN)
		env.Console.Secrets = append(env.Console.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-businesscentral-app-secret", cr.Name),
				Labels: map[string]string{
					"app":         cr.Name,
					"application": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(consoleCN, "jboss", []byte(cr.Spec.CommonConfig.KeyStorePassword)),
			},
		})
	}

	// server(s) keystore generation
	for i, server := range env.Servers {
		if server.Omit {
			break
		}
		serverCN := ""
		for _, rt := range server.Routes {
			if checkTLS(rt.Spec.TLS) {
				// use host of first tls route in env template
				serverCN = reconciler.GetRouteHost(rt, cr)
				break
			}
		}
		if serverCN == "" {
			serverCN = cr.Name
		}
		defaults.ConfigureHostname(&server, cr, serverCN)
		serverSet, relativeIndex := defaults.GetServerSet(cr, i)
		kieName := serverSet.Name
		kieIndex := defaults.GetKieIndex(&serverSet, relativeIndex)
		server.Secrets = append(server.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s%s-app-secret", kieName, kieIndex),
				Labels: map[string]string{
					"app":         cr.Name,
					"application": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(serverCN, "jboss", []byte(cr.Spec.CommonConfig.KeyStorePassword)),
			},
		})

		env.Servers[i] = server
	}

	// smartrouter keystore generation
	if !env.Smartrouter.Omit {
		smartCN := ""
		for _, rt := range env.Smartrouter.Routes {
			if checkTLS(rt.Spec.TLS) {
				// use host of first tls route in env template
				smartCN = reconciler.GetRouteHost(rt, cr)
				break
			}
		}
		if smartCN == "" {
			smartCN = cr.Name
		}

		defaults.ConfigureHostname(&env.Smartrouter, cr, smartCN)
		env.Smartrouter.Secrets = append(env.Smartrouter.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-smartrouter-app-secret", cr.Name),
				Labels: map[string]string{
					"app":         cr.Name,
					"application": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(smartCN, "jboss", []byte(cr.Spec.CommonConfig.KeyStorePassword)),
			},
		})
	}
	env = consolidateObjects(env, cr)

	rResult, err := reconciler.CreateCustomObjects(env.Console, cr)
	if err != nil {
		return env, rResult, err
	}
	rResult, err = reconciler.CreateCustomObjects(env.Smartrouter, cr)
	if err != nil {
		return env, rResult, err
	}
	for _, s := range env.Servers {
		rResult, err = reconciler.CreateCustomObjects(s, cr)
		if err != nil {
			return env, rResult, err
		}
	}
	for _, o := range env.Others {
		rResult, err = reconciler.CreateCustomObjects(o, cr)
		if err != nil {
			return env, rResult, err
		}
	}

	return env, rResult, nil
}

func consolidateObjects(env v1.Environment, cr *v1.KieApp) v1.Environment {
	env.Console = shared.ConstructObject(env.Console, cr.Spec.Objects.Console.KieAppObject)
	env.Smartrouter = shared.ConstructObject(env.Smartrouter, cr.Spec.Objects.Smartrouter)
	for index := range env.Servers {
		serverSet, _ := defaults.GetServerSet(cr, index)
		env.Servers[index] = shared.ConstructObject(env.Servers[index], serverSet.SecuredKieAppObject.KieAppObject)
	}
	return env
}

// CreateCustomObjects goes through all the different object types in the given CustomObject and creates them, if necessary
func (reconciler *Reconciler) CreateCustomObjects(object v1.CustomObject, cr *v1.KieApp) (reconcile.Result, error) {
	if object.Omit {
		return reconcile.Result{}, nil
	}
	var allObjects []v1.OpenShiftObject
	for index := range object.PersistentVolumeClaims {
		object.PersistentVolumeClaims[index].SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))
		allObjects = append(allObjects, &object.PersistentVolumeClaims[index])
	}
	for index := range object.ServiceAccounts {
		object.ServiceAccounts[index].SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		allObjects = append(allObjects, &object.ServiceAccounts[index])
	}
	for index := range object.Secrets {
		object.Secrets[index].SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
		allObjects = append(allObjects, &object.Secrets[index])
	}
	for index := range object.Roles {
		object.Roles[index].SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("Role"))
		allObjects = append(allObjects, &object.Roles[index])
	}
	for index := range object.RoleBindings {
		object.RoleBindings[index].SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("RoleBinding"))
		allObjects = append(allObjects, &object.RoleBindings[index])
	}
	for index := range object.DeploymentConfigs {
		object.DeploymentConfigs[index].SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		if len(object.BuildConfigs) == 0 {
			for ti, trigger := range object.DeploymentConfigs[index].Spec.Triggers {
				if trigger.Type == oappsv1.DeploymentTriggerOnImageChange {
					object.DeploymentConfigs[index].Spec.Triggers[ti].ImageChangeParams.From.Namespace, _ = reconciler.ensureImageStream(
						trigger.ImageChangeParams.From.Name,
						trigger.ImageChangeParams.From.Namespace,
						cr,
					)
				}
			}
		}
		allObjects = append(allObjects, &object.DeploymentConfigs[index])
	}
	for index := range object.Services {
		object.Services[index].SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		allObjects = append(allObjects, &object.Services[index])
	}
	for index := range object.Routes {
		object.Routes[index].SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
		allObjects = append(allObjects, &object.Routes[index])
	}
	for index := range object.ImageStreams {
		object.ImageStreams[index].SetGroupVersionKind(oimagev1.SchemeGroupVersion.WithKind("ImageStream"))
		allObjects = append(allObjects, &object.ImageStreams[index])
	}
	for index := range object.BuildConfigs {
		object.BuildConfigs[index].SetGroupVersionKind(buildv1.SchemeGroupVersion.WithKind("BuildConfig"))
		if object.BuildConfigs[index].Spec.Strategy.Type == buildv1.SourceBuildStrategyType {
			object.BuildConfigs[index].Spec.Strategy.SourceStrategy.From.Namespace, _ = reconciler.ensureImageStream(
				object.BuildConfigs[index].Spec.Strategy.SourceStrategy.From.Name,
				object.BuildConfigs[index].Spec.Strategy.SourceStrategy.From.Namespace, cr,
			)
		}
		allObjects = append(allObjects, &object.BuildConfigs[index])
	}

	for _, obj := range allObjects {
		_, err := reconciler.createCustomObject(obj, cr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (reconciler *Reconciler) ensureImageStream(name string, namespace string, cr *v1.KieApp) (string, error) {
	if reconciler.checkImageStreamTag(name, namespace) {
		return namespace, nil
	} else if reconciler.checkImageStreamTag(name, cr.Namespace) {
		return cr.Namespace, nil
	} else {
		log.Warnf("ImageStreamTag %s/%s doesn't exist.", namespace, name)
		err := reconciler.createLocalImageTag(name, cr)
		if err != nil {
			log.Error(err)
			return namespace, err
		}
		return cr.Namespace, nil
	}
}

// createCustomObject checks for an object's existence before creating it
func (reconciler *Reconciler) createCustomObject(obj v1.OpenShiftObject, cr *v1.KieApp) (reconcile.Result, error) {
	name := obj.GetName()
	namespace := cr.GetNamespace()
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", name, "namespace", namespace)

	err := controllerutil.SetControllerReference(cr, obj, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to create. ", err)
		return reconcile.Result{}, err
	}
	obj.SetNamespace(namespace)
	emptyObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	return reconciler.createObj(
		obj,
		reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, emptyObj),
	)
}

// createObj creates an object based on the error passed in from a `client.Get`
func (reconciler *Reconciler) createObj(obj v1.OpenShiftObject, err error) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())

	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		log.Info("Creating")
		err = reconciler.Service.Create(context.TODO(), obj)
		if err != nil {
			log.Warn("Failed to create object. ", err)
			return reconcile.Result{}, err
		}
		// Object created successfully - return and requeue
		return reconcile.Result{RequeueAfter: time.Duration(200) * time.Millisecond}, nil
	} else if err != nil {
		log.Error("Failed to get object. ", err)
		return reconcile.Result{}, err
	}
	log.Debug("Skip reconcile - object already exists")
	return reconcile.Result{}, nil
}

// UpdateObj reconciles the given object
func (reconciler *Reconciler) UpdateObj(obj v1.OpenShiftObject) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	log.Info("Updating")
	err := reconciler.Service.Update(context.TODO(), obj)
	if err != nil {
		log.Warn("Failed to update object. ", err)
		return reconcile.Result{}, err
	}
	// Object updated - return and requeue
	return reconcile.Result{Requeue: true}, nil
}

func checkTLS(tls *routev1.TLSConfig) bool {
	if tls != nil {
		return true
	}
	return false
}

// getPodNames returns the pod names of the array of pods passed in
func getDcNames(dcs []oappsv1.DeploymentConfig, cr *v1.KieApp) []string {
	var dcNames []string
	for _, dc := range dcs {
		for _, ownerRef := range dc.GetOwnerReferences() {
			if ownerRef.UID == cr.UID {
				dcNames = append(dcNames, dc.Name)
			}
		}
	}
	return dcNames
}

// GetRouteHost returns the Hostname of the route provided
func (reconciler *Reconciler) GetRouteHost(route routev1.Route, cr *v1.KieApp) string {
	route.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
	log := log.With("kind", route.GetObjectKind().GroupVersionKind().Kind, "name", route.Name, "namespace", route.Namespace)
	err := controllerutil.SetControllerReference(cr, &route, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Error setting controller reference. ", err)
	}
	route.SetNamespace(cr.Namespace)
	found := &routev1.Route{}
	err = reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		_, err = reconciler.createObj(
			&route,
			err,
		)
		if err != nil {
			log.Error("Error creating Route. ", err)
		}
	}

	found = &routev1.Route{}
	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		err = reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Error("Error getting Route. ", err)
	}

	return found.Spec.Host
}

// CreateConfigMaps generates & creates necessary ConfigMaps from embedded files
func (reconciler *Reconciler) CreateConfigMaps(myDep *appsv1.Deployment) {
	configMaps := defaults.ConfigMapsFromFile(myDep, myDep.Namespace, reconciler.Service.GetScheme())
	for _, configMap := range configMaps {
		var testDir bool
		result := strings.Split(configMap.Name, "-")
		if len(result) > 1 {
			if result[1] == "testdata" {
				testDir = true
			}
		}
		// don't create configmaps for test directories
		if !testDir {
			// if configmap already exists, compare to new
			if existingCM, exists := reconciler.createConfigMap(&configMap); exists {
				// if new configmap and existing have different data
				if !reflect.DeepEqual(configMap.Data, existingCM.Data) || !reflect.DeepEqual(configMap.BinaryData, existingCM.BinaryData) {
					log.Infof("Differences detected in %s ConfigMap.", configMap.Name)
					existingCM.Name = strings.Join([]string{configMap.Name, "bak"}, "-")
					for annotation, ver := range configMap.Annotations {
						if annotation == v1.SchemeGroupVersion.Group {
							existingCM.Name = strings.Join([]string{configMap.Name, ver, "bak"}, "-")
						}
					}
					existingCM.ResourceVersion = ""
					existingCM.OwnerReferences = nil
					// create a backup configmap of existing
					// if backup configmap already exists, overwrite w/ new backup
					if existingBackupCM, exists := reconciler.createConfigMap(existingCM); exists {
						// if backup configmap and existing backup have different data
						if !reflect.DeepEqual(existingCM.Data, existingBackupCM.Data) || !reflect.DeepEqual(existingCM.BinaryData, existingBackupCM.BinaryData) {
							existingBackupCM.Data = existingCM.Data
							reconciler.UpdateObj(existingBackupCM)
						}
					}
				}
			}
		}
	}
}

// createConfigMap creates an individual ConfigMap, will return the existing ConfigMap object should one exist
func (reconciler *Reconciler) createConfigMap(obj v1.OpenShiftObject) (*corev1.ConfigMap, bool) {
	emptyObj := &corev1.ConfigMap{}
	err := reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, emptyObj)
	if errors.IsNotFound(err) {
		// attempt creation of configmap if doesn't exist
		reconciler.createObj(obj, err)
		return &corev1.ConfigMap{}, false
	} else if err != nil {
		if err != nil {
			log.Error(err)
		}
		return &corev1.ConfigMap{}, false
	}
	return emptyObj, true
}
