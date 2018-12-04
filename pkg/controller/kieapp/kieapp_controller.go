package kieapp

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/kieserver"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	oappsv1 "github.com/openshift/api/apps/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"github.com/sirupsen/logrus"
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

// Add creates a new KieApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	imageClient, err := imagev1.NewForConfig(mgr.GetConfig())
	if err != nil {
		logrus.Error(err)
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
	logrus.Debugf("Reconciling %s/%s", request.Namespace, request.Name)

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
			deepCopyObj := configMap.DeepCopyObject()
			emptyObj := &corev1.ConfigMap{}
			err := reconciler.client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, emptyObj)
			if err != nil {
				_, err := reconciler.createObj(
					configMap.Name,
					configMap.Namespace,
					deepCopyObj,
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

	env, rResult, err := reconciler.NewEnv(instance)
	if err != nil {
		return rResult, err
	}

	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &oappsv1.DeploymentConfigList{}
	err = reconciler.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		logrus.Warnf("Failed to list dc's: %v", err)
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
			newDC := uDc.DeepCopyObject()
			logrus.Infof("Updating %s %s/%s", uDc.Kind, uDc.Namespace, uDc.Name)
			rResult, err := reconciler.updateObj(newDC)
			if err != nil {
				return rResult, err
			}
		}
		return rResult, nil
	}

	// Update status.Deployments if needed
	if !reflect.DeepEqual(dcNames, instance.Status.Deployments) {
		instance.Status.Deployments = dcNames
		return reconciler.updateObj(instance)
	}

	return rResult, nil
}

// Check ImageStream
func (reconciler *ReconcileKieApp) checkImageStreamTag(name, namespace string) bool {
	result := strings.Split(name, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	_, err := reconciler.imageClient.ImageStreamTags(namespace).Get(tagName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

// Create ImageStreamTag
func (reconciler *ReconcileKieApp) createLocalImageTag(currentTagReference corev1.ObjectReference, cr *v1.KieApp) error {
	result := strings.Split(currentTagReference.Name, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	version := []byte(cr.Spec.Template.Version)

	isnew := &oimagev1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tagName,
			Namespace: cr.Namespace,
		},
		Tag: &oimagev1.TagReference{
			Name: result[1],
			From: &corev1.ObjectReference{
				Kind: "DockerImage",
				Name: fmt.Sprintf("registry.access.redhat.com/rhpam-%s/%s", string(version[0]), tagName),
			},
		},
	}
	isnew.SetGroupVersionKind(oimagev1.SchemeGroupVersion.WithKind("ImageStreamTag"))

	logrus.Infof("Creating a new %s %s/%s", isnew.GetObjectKind().GroupVersionKind().Kind, isnew.Namespace, isnew.Name)
	_, err := reconciler.imageClient.ImageStreamTags(isnew.Namespace).Create(isnew)
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Errorf("Issue creating ImageStream %s/%s - %v", isnew.Namespace, isnew.Name, err)
		return err
	}
	return nil
}

func (reconciler *ReconcileKieApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1.KieApp) []oappsv1.DeploymentConfig {
	update := false
	cContainer := current.Spec.Template.Spec.Containers[0]
	nContainer := new.Spec.Template.Spec.Containers[0]

	if !shared.EnvVarCheck(cContainer.Env, nContainer.Env) {
		logrus.Debugf("Changes detected in %s/%s DeploymentConfig 'Env' config -\nOLD - %v\nNEW - %v", current.Namespace, current.Name, cContainer.Env, nContainer.Env)
		update = true
	}
	if !reflect.DeepEqual(cContainer.Resources, nContainer.Resources) {
		logrus.Debugf("Changes detected in %s/%s DeploymentConfig 'Resource' config -\nOLD - %v\nNEW - %v", current.Namespace, current.Name, cContainer.Resources, nContainer.Resources)
		update = true
	}

	if update {
		dcnew := new
		err := controllerutil.SetControllerReference(cr, &dcnew, reconciler.scheme)
		if err != nil {
			logrus.Errorf("Error setting controller reference for dc %s - %v", dcnew.Name, err)
		}
		dcnew.SetNamespace(current.Namespace)
		dcnew.SetResourceVersion(current.ResourceVersion)
		dcnew.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))

		dcUpdates = append(dcUpdates, dcnew)
	}
	return dcUpdates
}

func (reconciler *ReconcileKieApp) NewEnv(cr *v1.KieApp) (v1.Environment, reconcile.Result, error) {
	env, common, err := defaults.GetEnvironment(cr, reconciler.client)
	if err != nil {
		return v1.Environment{}, reconcile.Result{Requeue: true}, err
	}

	// console keystore generation
	consoleCN := ""
	for _, rt := range env.Console.Routes {
		if checkTLS(rt.Spec.TLS) {
			// use host of first tls route in env template
			consoleCN = reconciler.getRouteHost(rt, cr)
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

	// server(s) keystore generation
	for i, server := range env.Servers {
		serverCN := ""
		for _, rt := range server.Routes {
			if checkTLS(rt.Spec.TLS) {
				// use host of first tls route in env template
				serverCN = reconciler.getRouteHost(rt, cr)
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
	env = ConsolidateObjects(env, common, cr)
	rResult, err := reconciler.updateObj(cr)
	if err != nil {
		return env, rResult, err
	}
	rResult, err = reconciler.createCustomObjects(env.Console, cr)
	if err != nil {
		return env, rResult, err
	}
	for _, s := range env.Servers {
		rResult, err = reconciler.createCustomObjects(s, cr)
		if err != nil {
			return env, rResult, err
		}
	}
	for _, o := range env.Others {
		rResult, err = reconciler.createCustomObjects(o, cr)
		if err != nil {
			return env, rResult, err
		}
	}

	return env, rResult, nil
}

func ConsolidateObjects(env v1.Environment, common v1.KieAppSpec, cr *v1.KieApp) v1.Environment {
	env.Console = rhpamcentr.ConstructObject(env.Console, common, cr)
	for i, s := range env.Servers {
		s = kieserver.ConstructObject(s, common, cr)
		env.Servers[i] = s
	}
	return env
}

func (reconciler *ReconcileKieApp) createCustomObjects(object v1.CustomObject, cr *v1.KieApp) (reconcile.Result, error) {
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
	for index := range object.RoleBindings {
		object.RoleBindings[index].SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("RoleBinding"))
		allObjects = append(allObjects, &object.RoleBindings[index])
	}
	for index := range object.DeploymentConfigs {
		object.DeploymentConfigs[index].SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		for ti, trigger := range object.DeploymentConfigs[index].Spec.Triggers {
			if trigger.Type == oappsv1.DeploymentTriggerOnImageChange {
				if !reconciler.checkImageStreamTag(trigger.ImageChangeParams.From.Name, trigger.ImageChangeParams.From.Namespace) {
					if !reconciler.checkImageStreamTag(trigger.ImageChangeParams.From.Name, cr.Namespace) {
						logrus.Warnf("ImageStreamTag %s/%s doesn't exist", trigger.ImageChangeParams.From.Namespace, trigger.ImageChangeParams.From.Name)
						err := reconciler.createLocalImageTag(trigger.ImageChangeParams.From, cr)
						if err != nil {
							logrus.Error(err)
						} else {
							trigger.ImageChangeParams.From.Namespace = cr.Namespace
						}
					} else {
						trigger.ImageChangeParams.From.Namespace = cr.Namespace
					}
				}
				object.DeploymentConfigs[index].Spec.Triggers[ti] = trigger
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

	for _, obj := range allObjects {
		_, err := reconciler.createCustomObject(obj, cr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// createCustomObject checks for an object's existence before creating it
func (reconciler *ReconcileKieApp) createCustomObject(obj v1.OpenShiftObject, cr *v1.KieApp) (reconcile.Result, error) {
	name := obj.GetName()
	namespace := cr.GetNamespace()
	err := controllerutil.SetControllerReference(cr, obj, reconciler.scheme)
	if err != nil {
		logrus.Errorf("Failed to create new %s %s/%s: %v", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name, err)
		return reconcile.Result{}, err
	}
	obj.SetNamespace(namespace)
	deepCopyObj := obj.DeepCopyObject()
	emptyObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	return reconciler.createObj(
		name,
		namespace,
		deepCopyObj,
		reconciler.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, emptyObj),
	)
}

// createObj creates an object based on the error passed in from a `client.Get`
func (reconciler *ReconcileKieApp) createObj(name, namespace string, obj runtime.Object, err error) (reconcile.Result, error) {
	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		logrus.Infof("Creating a new %s %s/%s", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
		err = reconciler.client.Create(context.TODO(), obj)
		if err != nil {
			logrus.Warnf("Failed to create new %s %s/%s: %v", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name, err)
			return reconcile.Result{}, err
		}
		// Object created successfully - return and requeue
		return reconcile.Result{RequeueAfter: time.Duration(200) * time.Millisecond}, nil
	} else if err != nil {
		logrus.Infof("Failed to get %s %s/%s: %v", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name, err)
		return reconcile.Result{}, err
	}
	logrus.Debugf("Skip reconcile: %s %s/%s already exists", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
	return reconcile.Result{}, nil
}

func (reconciler *ReconcileKieApp) updateObj(obj runtime.Object) (reconcile.Result, error) {
	err := reconciler.client.Update(context.TODO(), obj)
	if err != nil {
		logrus.Warnf("Failed to update %s: %v", obj.GetObjectKind().GroupVersionKind().Kind, err)
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

func (reconciler *ReconcileKieApp) getRouteHost(route routev1.Route, cr *v1.KieApp) string {
	route.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
	err := controllerutil.SetControllerReference(cr, &route, reconciler.scheme)
	if err != nil {
		logrus.Errorf("Error setting controller reference for route %s - %v", route.Name, err)
	}
	route.SetNamespace(cr.Namespace)
	dRoute := route.DeepCopyObject()
	found := &routev1.Route{}
	err = reconciler.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		_, err = reconciler.createObj(
			route.Name,
			route.Namespace,
			dRoute,
			err,
		)
		if err != nil {
			logrus.Error(err)
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
		logrus.Error(err)
	}

	return found.Spec.Host
}
