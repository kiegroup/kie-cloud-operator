package kieapp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	appv1alpha1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/kieserver"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	return &ReconcileKieApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kieapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KieApp
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KieApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &oappsv1.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbac_v1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KieApp{},
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
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KieApp object and makes changes based on the state read
// and what is in the KieApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKieApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// logrus.Printf("Reconciling KieApp %s/%s\n", request.Namespace, request.Name)

	// Fetch the KieApp instance
	instance := &appv1alpha1.KieApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	env, rResult, err := r.NewEnv(instance)
	if err != nil {
		return rResult, err
	}

	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &oappsv1.DeploymentConfigList{}
	err = r.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		logrus.Printf("Failed to list dc's: %v", err)
		return reconcile.Result{}, err
	}
	dcNames := getDcNames(dcList.Items, instance)

	// Update DeploymentConfigs if needed
	var dcUpdates []oappsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		for _, cDc := range env.Console.DeploymentConfigs {
			if dc.Name == cDc.Name {
				dcUpdates = r.dcUpdateCheck(dc, cDc, dcUpdates, instance)
			}
		}
		for _, server := range env.Servers {
			for _, sDc := range server.DeploymentConfigs {
				if dc.Name == sDc.Name {
					dcUpdates = r.dcUpdateCheck(dc, sDc, dcUpdates, instance)
				}
			}
		}
		for _, other := range env.Others {
			for _, oDc := range other.DeploymentConfigs {
				if dc.Name == oDc.Name {
					dcUpdates = r.dcUpdateCheck(dc, oDc, dcUpdates, instance)
				}
			}
		}
	}
	if len(dcUpdates) > 0 {
		for _, uDc := range dcUpdates {
			newDC := uDc.DeepCopyObject()
			rResult, err := r.updateObj(uDc.Name, uDc.Namespace, newDC)
			if err != nil {
				return rResult, err
			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Update status.Deployments if needed
	if !reflect.DeepEqual(dcNames, instance.Status.Deployments) {
		instance.Status.Deployments = dcNames
		return r.crUpdate(instance)
	}

	return rResult, nil
}

func (r *ReconcileKieApp) crUpdate(cr *appv1alpha1.KieApp) (reconcile.Result, error) {
	err := r.client.Update(context.TODO(), cr)
	if err != nil {
		logrus.Printf("failed to update instance status: %v", err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileKieApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *appv1alpha1.KieApp) []oappsv1.DeploymentConfig {
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
		controllerutil.SetControllerReference(cr, &dcnew, r.scheme)
		dcnew.SetNamespace(current.Namespace)
		dcnew.SetResourceVersion(current.ResourceVersion)
		dcnew.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		dcUpdates = append(dcUpdates, dcnew)
	}
	return dcUpdates
}

func (r *ReconcileKieApp) NewEnv(cr *appv1alpha1.KieApp) (appv1alpha1.Environment, reconcile.Result, error) {
	env, common, err := defaults.GetEnvironment(cr)
	if err != nil {
		return appv1alpha1.Environment{}, reconcile.Result{}, err
	}
	//defer shared.Zeroing(password)

	// console keystore generation
	consoleCN := ""
	for _, rt := range env.Console.Routes {
		if CheckTLS(rt.Spec.TLS) {
			// use host of first tls route in env template
			consoleCN = r.getRouteHost(rt, cr)
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
				serverCN = r.getRouteHost(rt, cr)
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
	rResult, err := r.crUpdate(cr)
	if err != nil {
		return env, rResult, err
	}
	rResult, _ = r.createObjects(env.Console, cr)
	if err != nil {
		return env, rResult, err
	}
	for _, s := range env.Servers {
		rResult, _ = r.createObjects(s, cr)
		if err != nil {
			return env, rResult, err
		}
	}
	for _, o := range env.Others {
		rResult, _ = r.createObjects(o, cr)
		if err != nil {
			return env, rResult, err
		}
	}

	return env, rResult, nil
}

func ConsolidateObjects(env appv1alpha1.Environment, common appv1alpha1.KieAppSpec, cr *appv1alpha1.KieApp) appv1alpha1.Environment {
	env.Console = rhpamcentr.ConstructObject(env.Console, common, cr)
	for i, s := range env.Servers {
		s = kieserver.ConstructObject(s, common, cr)
		env.Servers[i] = s
	}
	return env
}

func (r *ReconcileKieApp) createObjects(object appv1alpha1.CustomObject, cr *appv1alpha1.KieApp) (reconcile.Result, error) {
	for _, obj := range object.PersistentVolumeClaims {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"))
		dObj := obj.DeepCopyObject()
		found := &corev1.PersistentVolumeClaim{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.ServiceAccounts {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		dObj := obj.DeepCopyObject()
		found := &corev1.ServiceAccount{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.Secrets {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
		dObj := obj.DeepCopyObject()
		found := &corev1.Secret{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.RoleBindings {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(rbac_v1.SchemeGroupVersion.WithKind("RoleBinding"))
		dObj := obj.DeepCopyObject()
		found := &rbac_v1.RoleBinding{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.DeploymentConfigs {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
		dObj := obj.DeepCopyObject()
		found := &oappsv1.DeploymentConfig{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.Services {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		dObj := obj.DeepCopyObject()
		found := &corev1.Service{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}
	for _, obj := range object.Routes {
		controllerutil.SetControllerReference(cr, &obj, r.scheme)
		obj.SetNamespace(cr.Namespace)
		obj.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
		dObj := obj.DeepCopyObject()
		found := &routev1.Route{}
		_, _ = r.createObj(
			obj.Name,
			obj.Namespace,
			dObj,
			r.client.Get(context.TODO(), types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, found),
		)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileKieApp) createObj(name, namespace string, obj runtime.Object, err error) (reconcile.Result, error) {
	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		logrus.Printf("Creating a new %s %s/%s\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
		err = r.client.Create(context.TODO(), obj)
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

func (r *ReconcileKieApp) updateObj(name, namespace string, obj runtime.Object) (reconcile.Result, error) {
	logrus.Printf("Updating %s %s/%s\n", obj.GetObjectKind().GroupVersionKind().Kind, namespace, name)
	err := r.client.Update(context.TODO(), obj)
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
func getDcNames(dcs []oappsv1.DeploymentConfig, cr *appv1alpha1.KieApp) []string {
	var dcNames []string
	for _, dc := range dcs {
		for _, or := range dc.GetOwnerReferences() {
			if or.UID == cr.UID {
				dcNames = append(dcNames, dc.Name)
			}
		}
	}
	return dcNames
}

func (r *ReconcileKieApp) getRouteHost(route routev1.Route, cr *appv1alpha1.KieApp) string {
	controllerutil.SetControllerReference(cr, &route, r.scheme)
	route.SetNamespace(cr.Namespace)
	route.SetGroupVersionKind(routev1.SchemeGroupVersion.WithKind("Route"))
	dRoute := route.DeepCopyObject()
	found := &routev1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		_, err = r.createObj(
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
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
		if err == nil {
			break
		}
	}
	if err != nil {
		logrus.Error(err)
	}

	return found.Spec.Host
}
