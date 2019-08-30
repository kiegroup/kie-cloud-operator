package kieapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"os"
	"reflect"
	"strings"
	"time"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var name = "console-cr-form"
var operatorName string

func shouldDeployConsole() bool {
	shouldDeploy := os.Getenv(constants.OpUIEnv)
	if strings.ToLower(shouldDeploy) == "false" {
		log.Debugf("Environment variable %s set to %s, so will not deploy operator UI", constants.OpUIEnv, shouldDeploy)
		return false
	}
	//Default to deploying, if env var not set to false
	return true
}

func deployConsole(reconciler *Reconciler, operator *appsv1.Deployment) {
	log.Debugf("Checking operator-ui deployment")
	namespace := os.Getenv(constants.NameSpaceEnv)
	operatorName = os.Getenv(constants.OpNameEnv)
	role := getRole(namespace)
	roleBinding := getRoleBinding(namespace)
	sa := getServiceAccount(namespace)
	image := getImage(operator)
	pod := getPod(namespace, image, sa.Name, operator)
	service := getService(namespace)
	route := getRoute(namespace)
	requested := compare.NewMapBuilder().Add(role, roleBinding, sa, pod, service, route).ResourceMap()
	deployed, err := loadCounterparts(reconciler, requested)
	if err != nil {
		log.Error("Failed to load deployed resources.", err)
		return
	}
	comparator := compare.NewMapComparator()
	deltas := comparator.Compare(deployed, requested)
	writer := write.New().WithOwnerReferences(operator.GetOwnerReferences()...)
	var hasUpdates bool
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Debugf("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
		added, err := writer.AddResources(reconciler.Service.GetScheme(), reconciler.Service, delta.Added)
		if err != nil {
			log.Warnf("Got error applying changes %v", err)
			return
		}
		updated, err := writer.UpdateResources(deployed[resourceType], reconciler.Service.GetScheme(), reconciler.Service, delta.Updated)
		if err != nil {
			log.Warnf("Got error applying changes %v", err)
			return
		}
		removed, err := writer.RemoveResources(reconciler.Service, delta.Removed)
		if err != nil {
			log.Warnf("Got error applying changes %v", err)
			return
		}
		hasUpdates = hasUpdates || added || updated || removed
	}
	updateCSVlinks(reconciler, route, operator)
}

func loadCounterparts(reconciler *Reconciler, requestedMap map[reflect.Type][]resource.KubernetesResource) (map[reflect.Type][]resource.KubernetesResource, error) {
	deployedMap := make(map[reflect.Type][]resource.KubernetesResource)
	for resourceType, requestedArray := range requestedMap {
		var deployedArray []resource.KubernetesResource
		for _, requested := range requestedArray {
			deployed := reflect.New(resourceType).Interface().(resource.KubernetesResource)
			err := reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: requested.GetName(), Namespace: requested.GetNamespace()}, deployed)
			if err == nil {
				deployedArray = append(deployedArray, deployed)
			} else {
				if errors.IsNotFound(err) {
					//Not found, just don't add it to array
				} else {
					return nil, err
				}
			}
		}
		deployedMap[resourceType] = deployedArray
	}
	return deployedMap, nil
}

func updateCSVlinks(reconciler *Reconciler, route *routev1.Route, operator *appsv1.Deployment) {
	var update bool
	found := &routev1.Route{}
	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		err := reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
		if err == nil && found.Spec.Host != route.Name && found.Spec.Host != "" {
			break
		}
	}
	if found.Name == "" || found.Spec.Host == found.Name {
		log.Warn("Unable to get Route url. ", route.Name)
		return
	}
	url := fmt.Sprintf("https://%s", found.Spec.Host)
	csv := reconciler.getCSV(operator)
	if reflect.DeepEqual(csv, &operatorsv1alpha1.ClusterServiceVersion{}) {
		log.Info("No ClusterServiceVersion found, likely because no such owner was set on the operator. This might be because the operator was not installed through OLM")
		return
	}
	link := getConsoleLink(csv)
	if link == nil || link.URL != url {
		update = true
		if link == nil {
			csv.Spec.Links = append([]operatorsv1alpha1.AppLink{{Name: constants.ConsoleLinkName, URL: url}}, csv.Spec.Links...)
		} else {
			link.URL = url
		}
	}
	if !strings.Contains(csv.Spec.Description, constants.ConsoleDescription) {
		update = true
		csv.Spec.Description = strings.Join([]string{csv.Spec.Description, constants.ConsoleDescription}, "\n\n")
	}
	if update {
		log.Debugf("Updating ", csv.Name, " ", csv.Kind, ".")
		err := reconciler.Service.Update(context.TODO(), csv)
		if err != nil {
			log.Error("Failed to update CSV. ", err)
		}
	}
}

func getConsoleLink(csv *operatorsv1alpha1.ClusterServiceVersion) *operatorsv1alpha1.AppLink {
	for i, link := range csv.Spec.Links {
		if link.Name == constants.ConsoleLinkName {
			return &csv.Spec.Links[i]
		}
	}
	return nil
}

func getPod(namespace string, image string, sa string, operator *appsv1.Deployment) *corev1.Pod {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	volume := corev1.Volume{Name: "proxy-tls"}
	volume.Secret = &corev1.SecretVolumeSource{SecretName: volume.Name}
	sar, err := json.Marshal(map[string]string{
		"namespace": namespace,
		"resource":  "kieapps",
		"name":      name,
		"verb":      "create",
	})
	if err != nil {
		log.Error("Failed to marshal sar config to json. ", err)
	}
	debug := constants.DebugFalse
	if shared.EnvVarSet(constants.DebugTrue, operator.Spec.Template.Spec.Containers[0].Env) {
		debug = constants.DebugTrue
	}
	sarString := fmt.Sprintf("--openshift-sar=%s", sar)
	httpPort := int32(8080)
	httpsPort := int32(8443)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: sa,
			Volumes:            []corev1.Volume{volume},
			Containers: []corev1.Container{
				{
					Name:  "oauth-proxy",
					Image: "registry.access.redhat.com/openshift3/oauth-proxy",
					Ports: []corev1.ContainerPort{{Name: "public", ContainerPort: httpsPort}},
					Args: []string{
						"--http-address=",
						fmt.Sprintf("--https-address=:%d", httpsPort),
						fmt.Sprintf("--upstream=http://localhost:%d", httpPort),
						"--provider=openshift",
						sarString,
						fmt.Sprintf("--openshift-service-account=%s", sa),
						"--tls-cert=/etc/tls/private/tls.crt",
						"--tls-key=/etc/tls/private/tls.key",
						"--cookie-secret=SECRET",
					},
					VolumeMounts: []corev1.VolumeMount{{Name: volume.Name, MountPath: "/etc/tls/private"}},
				},
				{
					Name:            name,
					Image:           image,
					ImagePullPolicy: operator.Spec.Template.Spec.Containers[0].ImagePullPolicy,
					Command: []string{
						"console-cr-form",
					},
					Ports: []corev1.ContainerPort{{Name: "http", ContainerPort: httpPort}},
					Env:   []corev1.EnvVar{debug},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: httpPort},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       3,
						FailureThreshold:    20,
					},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.IntOrString{IntVal: httpPort},
							},
						},
						InitialDelaySeconds: 60,
						PeriodSeconds:       60,
					},
				},
			},
		},
	}
}

func getImage(operator *appsv1.Deployment) string {
	image := operator.Spec.Template.Spec.Containers[0].Image
	return image
}

func getService(namespace string) *corev1.Service {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"service.alpha.openshift.io/serving-cert-secret-name": "proxy-tls",
			},
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					TargetPort: intstr.FromInt(8443),
					Name:       "proxy",
				},
			},
			Selector: labels,
		},
	}
}

func getRoute(namespace string) *routev1.Route {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationReencrypt,
			},
		},
	}
}

func getRole(namespace string) *rbacv1.Role {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{api.SchemeGroupVersion.Group},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}
}

func getRoleBinding(namespace string) *rbacv1.RoleBinding {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: name,
			},
		},
	}
}

func getServiceAccount(namespace string) *corev1.ServiceAccount {
	labels := map[string]string{
		"app":  operatorName,
		"name": name,
	}
	type reference struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
	}
	type annotationType struct {
		Kind       string    `json:"kind"`
		APIVersion string    `json:"apiVersion"`
		Reference  reference `json:"reference"`
	}
	annotation, err := json.Marshal(annotationType{
		Kind:       "OAuthRedirectReference",
		APIVersion: "v1",
		Reference: reference{
			Kind: "Route",
			Name: name,
		},
	})
	if err != nil {
		log.Error("Failed to marshal annotation config to json. ", err)
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.primary": string(annotation),
			},
		},
	}
}
