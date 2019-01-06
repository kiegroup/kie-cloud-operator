package kieapp

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logs.GetLogger("kieapp.controller")

// Add creates a new KieApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	imageClient, err := imagev1.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error("Error getting image client. ", err)
	}
	return &ReconcileKieApp{client: mgr.GetClient(), cache: mgr.GetCache(), imageClient: imageClient, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kieapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KieApp
	err = c.Watch(&source.Kind{Type: &v1.KieApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &oappsv1.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &buildv1.BuildConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &oimagev1.ImageStream{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1.KieApp{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKieApp{}

// ReconcileKieApp reconciles a KieApp object
type ReconcileKieApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	// can cache be leveraged instead in some locations?
	cache       cache.Cache
	imageClient *imagev1.ImageV1Client
	scheme      *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KieApp object and makes changes based on the state read
// and what is in the KieApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (reconciler *ReconcileKieApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Create critical ConfigMaps if don't exist
	configMaps := defaults.ConfigMapsFromFile(request.Namespace)
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
			configMap.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
			emptyObj := &corev1.ConfigMap{}
			err := reconciler.client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, emptyObj)
			if err != nil {
				_, err := reconciler.createObj(
					&configMap,
					err,
				)
				if err != nil && !errors.IsAlreadyExists(err) {
					return reconcile.Result{RequeueAfter: time.Duration(1) * time.Second}, err
				}
			}
		}
	}

	// Fetch the KieApp instance
	instance := &v1.KieApp{}
	err := reconciler.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)

	env, rResult, err := NewEnv(reconciler, instance)
	if err != nil {
		return rResult, err
	}

	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &oappsv1.DeploymentConfigList{}
	err = reconciler.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		log.Warn("Failed to list dc's. ", err)
		return reconcile.Result{}, err
	}
	dcNames := getDcNames(dcList.Items, instance)

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
				return rResult, err
			}
		}
		return rResult, nil
	}

	// Fetch the cached KieApp instance
	cachedInstance := &v1.KieApp{}
	err = reconciler.cache.Get(context.TODO(), request.NamespacedName, cachedInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	instance.Status.Deployments = dcNames
	// Update CR if needed
	if !reflect.DeepEqual(instance, cachedInstance) {
		return reconciler.UpdateObj(instance)
	}

	return reconcile.Result{}, nil
}

func (reconciler *ReconcileKieApp) GetClient() client.Client {
	return reconciler.client
}

// Check ImageStream
func (reconciler *ReconcileKieApp) checkImageStreamTag(name, namespace string) bool {
	log := log.With("kind", "ImageStreamTag", "name", name, "namespace", namespace)
	result := strings.Split(name, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	_, err := reconciler.imageClient.ImageStreamTags(namespace).Get(tagName, metav1.GetOptions{})
	if err != nil {
		log.Debug("Object does not exist")
		return false
	}
	return true
}

// Create local ImageStreamTag
func (reconciler *ReconcileKieApp) createLocalImageTag(tagRefName string, cr *v1.KieApp) error {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	version := []byte(cr.Spec.Template.Version)
	imageName := tagName
	regContext := fmt.Sprintf("rhpam-%s", string(version[0]))

	registryAddress := cr.Spec.RhpamRegistry.Registry
	if strings.Contains(result[0], "businesscentral-indexing-openshift") {
		regContext = "rhpam-7-tech-preview"
	} else if strings.Contains(result[0], "amq-broker-7") {
		registryAddress = constants.RhpamRegistry
		regContext = "amq-broker-7"
	} else if result[0] == "postgresql" || result[0] == "mysql" {
		registryAddress = constants.RhpamRegistry
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
		},
	}
	isnew.SetGroupVersionKind(oimagev1.SchemeGroupVersion.WithKind("ImageStreamTag"))
	if cr.Spec.RhpamRegistry.Insecure {
		isnew.Tag.ImportPolicy = oimagev1.TagImportPolicy{
			Insecure: true,
		}
	}
	log := log.With("kind", isnew.GetObjectKind().GroupVersionKind().Kind, "name", isnew.Name, "from", isnew.Tag.From.Name, "namespace", isnew.Namespace)
	log.Info("Creating")
	_, err := reconciler.imageClient.ImageStreamTags(isnew.Namespace).Create(isnew)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error("Issue creating object. ", err)
		return err
	}
	return nil
}

func (reconciler *ReconcileKieApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1.KieApp) []oappsv1.DeploymentConfig {
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
		err := controllerutil.SetControllerReference(cr, &dcnew, reconciler.scheme)
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

func NewEnv(reconciler v1.PlatformService, cr *v1.KieApp) (v1.Environment, reconcile.Result, error) {
	env, err := defaults.GetEnvironment(cr, reconciler.GetClient())
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
				break
			}
		}
		if consoleCN == "" {
			consoleCN = cr.Name
		}
		env.Console.Secrets = append(env.Console.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-businesscentral-app-secret", cr.Name),
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(consoleCN, "jboss", []byte(cr.Spec.Template.KeyStorePassword)),
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
		server.Secrets = append(server.Secrets, corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-kieserver-%d-app-secret", cr.Name, i),
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Data: map[string][]byte{
				"keystore.jks": shared.GenerateKeystore(serverCN, "jboss", []byte(cr.Spec.Template.KeyStorePassword)),
			},
		})

		env.Servers[i] = server
	}
	env = ConsolidateObjects(env, cr)

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

func ConsolidateObjects(env v1.Environment, cr *v1.KieApp) v1.Environment {
	env.Console = shared.ConstructObject(env.Console, &cr.Spec.Objects.Console)
	env.Smartrouter = shared.ConstructObject(env.Smartrouter, &cr.Spec.Objects.Smartrouter)
	for i, s := range env.Servers {
		s = shared.ConstructObject(s, &cr.Spec.Objects.Server)
		env.Servers[i] = s
	}
	return env
}

func (reconciler *ReconcileKieApp) CreateCustomObjects(object v1.CustomObject, cr *v1.KieApp) (reconcile.Result, error) {
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
					namespace, err := reconciler.ensureImageStream(trigger.ImageChangeParams.From.Name, trigger.ImageChangeParams.From.Namespace, cr)
					if err == nil {
						object.DeploymentConfigs[index].Spec.Triggers[ti].ImageChangeParams.From.Namespace = namespace
					}
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
			from := object.BuildConfigs[index].Spec.Strategy.SourceStrategy.From
			namespace, err := reconciler.ensureImageStream(from.Name, from.Namespace, cr)
			if err == nil {
				from.Namespace = namespace
			}
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

func (reconciler *ReconcileKieApp) ensureImageStream(name string, namespace string, cr *v1.KieApp) (string, error) {
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
func (reconciler *ReconcileKieApp) createCustomObject(obj v1.OpenShiftObject, cr *v1.KieApp) (reconcile.Result, error) {
	name := obj.GetName()
	namespace := cr.GetNamespace()
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", name, "namespace", namespace)

	err := controllerutil.SetControllerReference(cr, obj, reconciler.scheme)
	if err != nil {
		log.Error("Failed to create. ", err)
		return reconcile.Result{}, err
	}
	obj.SetNamespace(namespace)
	emptyObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	return reconciler.createObj(
		obj,
		reconciler.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, emptyObj),
	)
}

// createObj creates an object based on the error passed in from a `client.Get`
func (reconciler *ReconcileKieApp) createObj(obj v1.OpenShiftObject, err error) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())

	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		log.Info("Creating")
		err = reconciler.client.Create(context.TODO(), obj)
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

func (reconciler *ReconcileKieApp) UpdateObj(obj v1.OpenShiftObject) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	log.Info("Updating")
	err := reconciler.client.Update(context.TODO(), obj)
	if err != nil {
		log.Warn("Failed to update object. ", err)
		return reconcile.Result{}, err
	}
	// Spec updated - return and requeue
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

func (reconciler *ReconcileKieApp) GetRouteHost(route routev1.Route, cr *v1.KieApp) string {
	route.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
	log := log.With("kind", route.GetObjectKind().GroupVersionKind().Kind, "name", route.Name, "namespace", route.Namespace)
	err := controllerutil.SetControllerReference(cr, &route, reconciler.scheme)
	if err != nil {
		log.Error("Error setting controller reference. ", err)
	}
	route.SetNamespace(cr.Namespace)
	found := &routev1.Route{}
	err = reconciler.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
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
		err = reconciler.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Error("Error getting Route. ", err)
	}

	return found.Spec.Host
}
