package kieapp

import (
	"context"
	"fmt"
	"log"
	"reflect"

	appv1alpha1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1alpha1"
	oappsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKieApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling KieApp %s/%s\n", request.Namespace, request.Name)

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

	// Check if the DeploymentConfig already exists, if not create a new one
	founddc := &oappsv1.DeploymentConfig{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, founddc)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DeploymentConfig
		dc := r.deploymentConfigForCR(instance)
		log.Printf("Creating a new DeploymentConfig %s/%s\n", dc.Namespace, dc.Name)
		err = r.client.Create(context.TODO(), dc)
		if err != nil {
			log.Printf("Failed to create new DeploymentConfig : %v\n", err)
			return reconcile.Result{}, err
		}
		// DeploymentConfig created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Printf("Failed to get DeploymentConfig : %v\n", err)
		return reconcile.Result{}, err
	}

	// Ensure the DeploymentConfig size is the same as the spec
	size := instance.Spec.Size
	if founddc.Spec.Replicas != size {
		founddc.Spec.Replicas = size
		err = r.client.Update(context.TODO(), founddc)
		if err != nil {
			log.Printf("Failed to update DeploymentConfig : %v\n", err)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's DeploymentConfig
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForKieApp(instance.Name))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return reconcile.Result{}, err
	}

	dcList := &oappsv1.DeploymentConfigList{}
	err = r.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		log.Printf("Failed to list dc's: %v", err)
		return reconcile.Result{}, err
	}
	dcNames := getDcNames(dcList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(dcNames, instance.Status.Nodes) {
		instance.Status.Nodes = dcNames
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			log.Printf("failed to update instance status: %v", err)
			return reconcile.Result{}, err
		}
		// return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(5) * time.Second}, nil
	}

	// DC already exists - don't requeue
	log.Printf("Skip reconcile: DC %s/%s already exists", founddc.Namespace, founddc.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileKieApp) deploymentConfigForCR(cr *appv1alpha1.KieApp) *oappsv1.DeploymentConfig {
	ls := labelsForKieApp(cr.Name)
	replicas := cr.Spec.Size

	dc := &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    ls,
		},
		Spec: oappsv1.DeploymentConfigSpec{
			Replicas: replicas,
			Selector: ls,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "busybox",
						Image:   "busybox",
						Command: []string{"sleep", "36000"},
					}},
				},
			},
		},
	}

	fmt.Println(dc)
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(cr, dc, r.scheme)
	return dc
}

func labelsForKieApp(name string) map[string]string {
	return map[string]string{"app": "rhpam", "rhpam_cr": name}
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
func getDcNames(dcs []oappsv1.DeploymentConfig) []string {
	var dcNames []string
	for _, dc := range dcs {
		dcNames = append(dcNames, dc.Name)
	}
	return dcNames
}
