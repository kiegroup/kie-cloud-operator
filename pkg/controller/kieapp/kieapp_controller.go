package kieapp

import (
	"context"
	"fmt"
	"reflect"
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
	rbac_v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
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
	return &ReconcileKieApp{client: mgr.GetClient(), clientConfig: mgr.GetConfig(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kieapp-controller", mgr, controller.Options{Reconciler: r})
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

	err = c.Watch(&source.Kind{Type: &rbac_v1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
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
	client       client.Client
	clientConfig *rest.Config
	scheme       *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KieApp object and makes changes based on the state read
// and what is in the KieApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (reconciler *ReconcileKieApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// logrus.Printf("Reconciling KieApp %s/%s\n", request.Namespace, request.Name)

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
		logrus.Printf("Failed to list dc's: %v", err)
		return reconcile.Result{}, err
	}
	dcNames := getDcNames(dcList.Items, instance)

	// Update DeploymentConfigs if needed
	var dcUpdates []oappsv1.DeploymentConfig
	isTags := map[string]string{"centos:7": "openshift"}
	//isTags := make(map[string]string)
	for _, dc := range dcList.Items {
		for _, cDc := range env.Console.DeploymentConfigs {
			for _, trigger := range cDc.Spec.Triggers {
				if trigger.Type == oappsv1.DeploymentTriggerOnImageChange {
					if !existsInMap(isTags, trigger.ImageChangeParams.From.Name) {
						isTags[trigger.ImageChangeParams.From.Name] = trigger.ImageChangeParams.From.Namespace
					}
				}
			}
			if dc.Name == cDc.Name {
				dcUpdates = reconciler.dcUpdateCheck(dc, cDc, dcUpdates, instance)
			}
		}
		for _, server := range env.Servers {
			for _, sDc := range server.DeploymentConfigs {
				for _, trigger := range sDc.Spec.Triggers {
					if trigger.Type == oappsv1.DeploymentTriggerOnImageChange {
						if !existsInMap(isTags, trigger.ImageChangeParams.From.Name) {
							isTags[trigger.ImageChangeParams.From.Name] = trigger.ImageChangeParams.From.Namespace
						}
					}
				}
				if dc.Name == sDc.Name {
					dcUpdates = reconciler.dcUpdateCheck(dc, sDc, dcUpdates, instance)
				}
			}
		}
		for _, other := range env.Others {
			for _, oDc := range other.DeploymentConfigs {
				for _, trigger := range oDc.Spec.Triggers {
					if trigger.Type == oappsv1.DeploymentTriggerOnImageChange {
						if !existsInMap(isTags, trigger.ImageChangeParams.From.Name) {
							isTags[trigger.ImageChangeParams.From.Name] = trigger.ImageChangeParams.From.Namespace
						}
					}
				}
				if dc.Name == oDc.Name {
					dcUpdates = reconciler.dcUpdateCheck(dc, oDc, dcUpdates, instance)
				}
			}
		}
	}
	if len(dcUpdates) > 0 {
		for _, uDc := range dcUpdates {
			newDC := uDc.DeepCopyObject()
			rResult, err := reconciler.updateObj(uDc.Name, uDc.Namespace, newDC)
			if err != nil {
				return rResult, err
			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Check/Create ImageStream
	is := &oimagev1.ImageStreamTag{}
	imageV1Client, err := imagev1.NewForConfig(reconciler.clientConfig)
	var createIsTags []string
	if err != nil {
		return reconcile.Result{}, err
	}

	for image, ns := range isTags {
		is, err = imageV1Client.ImageStreamTags(ns).Get(image, metav1.GetOptions{})
		if err != nil {
			isTags[image] = instance.Namespace
			is, err = imageV1Client.ImageStreamTags(isTags[image]).Get(image, metav1.GetOptions{})
			if err != nil {
				logrus.Printf("Failed to get imagestream: %v", err)
				// !!!!!!!!!!!!!! how handle????
				createIsTags = append(createIsTags, image)
				return reconcile.Result{}, err
			}
		}
		if is != nil {
			logrus.Printf("%v/%v - %v", is.Namespace, is.Name, is.Image.Name)
		}
	}

	// Update status.Deployments if needed
	if !reflect.DeepEqual(dcNames, instance.Status.Deployments) {
		instance.Status.Deployments = dcNames
		return reconciler.crUpdate(instance)
	}

	return rResult, nil
}

func (reconciler *ReconcileKieApp) crUpdate(cr *v1.KieApp) (reconcile.Result, error) {
	err := reconciler.client.Update(context.TODO(), cr)
	if err != nil {
		logrus.Printf("failed to update instance status: %v", err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func (reconciler *ReconcileKieApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1.KieApp) []oappsv1.DeploymentConfig {
	update := false
	cContainer := current.Spec.Template.Spec.Containers[0]
	nContainer := new.Spec.Template.Spec.Containers[0]

	if !shared.EnvVarCheck(cContainer.Env, nContainer.Env) {
		update = true
	}
	if !reflect.DeepEqual(cContainer.Resources, nContainer.Resources) {
		update = true
	}
	if update {
		dcnew := new
		controllerutil.SetControllerReference(cr, &dcnew, reconciler.scheme)
		dcnew.SetNamespace(current.Namespace)
		dcnew.SetResourceVersion(current.ResourceVersion)
		dcnew.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		dcUpdates = append(dcUpdates, dcnew)
	}
	return dcUpdates
}

func (reconciler *ReconcileKieApp) NewEnv(cr *v1.KieApp) (v1.Environment, reconcile.Result, error) {
	env, common, err := defaults.GetEnvironment(cr)
	if err != nil {
		return v1.Environment{}, reconcile.Result{}, err
	}

	// console keystore generation
	consoleCN := ""
	for _, rt := range env.Console.Routes {
		if CheckTLS(rt.Spec.TLS) {
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
			if CheckTLS(rt.Spec.TLS) {
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
	rResult, err := reconciler.crUpdate(cr)
	if err != nil {
		return env, rResult, err
	}
	rResult, _ = reconciler.createCustomObjects(env.Console, cr)
	if err != nil {
		return env, rResult, err
	}
	for _, s := range env.Servers {
		rResult, _ = reconciler.createCustomObjects(s, cr)
		if err != nil {
			return env, rResult, err
		}
	}
	for _, o := range env.Others {
		rResult, _ = reconciler.createCustomObjects(o, cr)
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
	for _, obj := range object.PersistentVolumeClaims {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &corev1.PersistentVolumeClaim{})
	}
	for _, obj := range object.ServiceAccounts {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &corev1.ServiceAccount{})
	}
	for _, obj := range object.Secrets {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &corev1.Secret{})
	}
	for _, obj := range object.RoleBindings {
		obj.SetGroupVersionKind(rbac_v1.SchemeGroupVersion.WithKind("RoleBinding"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &rbac_v1.RoleBinding{})
	}
	for _, obj := range object.DeploymentConfigs {
		obj.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &oappsv1.DeploymentConfig{})
	}
	for _, obj := range object.Services {
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &corev1.Service{})
	}
	for _, obj := range object.Routes {
		obj.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
		controllerutil.SetControllerReference(cr, &obj, reconciler.scheme)
		obj.SetNamespace(cr.Namespace)
		_, _ = reconciler.createCustomObject(obj.Name, obj.Namespace, obj.DeepCopyObject(), &routev1.Route{})
	}

	return reconcile.Result{}, nil
}

// createCustomObject checks for an object's existence before creating it
func (reconciler *ReconcileKieApp) createCustomObject(name, namespace string, deepCopyObj, emptyObj runtime.Object) (reconcile.Result, error) {
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
		logrus.Printf("Creating a new %s %s/%s\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
		err = reconciler.client.Create(context.TODO(), obj)
		if err != nil {
			logrus.Printf("Failed to create new %s %s/%s: %v\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name, err)
			return reconcile.Result{}, err
		}
		// Object created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		logrus.Printf("Failed to get %s %s/%s: %v\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name, err)
		return reconcile.Result{}, err
	}
	// logrus.Printf("Skip reconcile: %s %s/%s already exists", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
	return reconcile.Result{}, nil
}

func (reconciler *ReconcileKieApp) updateObj(name, namespace string, obj runtime.Object) (reconcile.Result, error) {
	logrus.Printf("Updating %s %s/%s\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
	err := reconciler.client.Update(context.TODO(), obj)
	if err != nil {
		logrus.Printf("Failed to update %s : %v\n", obj.GetObjectKind().GroupVersionKind().Kind, err)
		return reconcile.Result{}, err
	}
	// Spec updated - return and requeue
	return reconcile.Result{Requeue: true}, nil
}

func CheckTLS(tls *routev1.TLSConfig) bool {
	if tls != nil {
		return true
	}
	return false
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
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
	controllerutil.SetControllerReference(cr, &route, reconciler.scheme)
	route.SetNamespace(cr.Namespace)
	dRoute := route.DeepCopyObject()
	found := &routev1.Route{}
	err := reconciler.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
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

func existsInMap(m map[string]string, i string) bool {
	for _, ele := range m {
		_, ok := m[ele]
		if ok {
			return true
		}
	}
	return false
}
