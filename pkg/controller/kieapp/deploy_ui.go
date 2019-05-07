package kieapp

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	rbacv1 "k8s.io/api/rbac/v1"
	"os"
	"strings"

	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var name = "console-cr-form"
var operatorName string

func shouldDeployConsole() bool {
	shouldDeploy := os.Getenv(constants.OpUiEnv)
	if strings.ToLower(shouldDeploy) == "false" {
		log.Debugf("Environment variable %s set to %s, so will not deploy operator UI", constants.OpUiEnv, shouldDeploy)
		return false
	} else {
		//Default to deploying, if env var not set to false
		return true
	}
}

func deployConsole(reconciler *Reconciler, operator *appsv1.Deployment) {
	log.Debugf("Will deploy operator-ui")
	namespace := os.Getenv(constants.NameSpaceEnv)
	operatorName = os.Getenv(constants.OpNameEnv)
	role := getRole(namespace)
	err := controllerutil.SetControllerReference(operator, role, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for role. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), role)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create role as it already exists", role)
		} else {
			log.Error("Failed to create role. ", err)
			return
		}
	}
	roleBinding := getRoleBinding(namespace)
	err = controllerutil.SetControllerReference(operator, roleBinding, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for roleBinding. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), roleBinding)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create roleBinding as it already exists", roleBinding)
		} else {
			log.Error("Failed to create roleBinding. ", err)
			return
		}
	}
	sa := getServiceAccount(namespace)
	err = controllerutil.SetControllerReference(operator, sa, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for serviceaccount. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), sa)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create serviceaccount as it already exists", sa)
		} else {
			log.Error("Failed to create serviceaccount. ", err)
			return
		}
	}
	image := getImage(operator)
	pod := getPod(namespace, image, sa.Name, operator)
	err = controllerutil.SetControllerReference(operator, pod, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for pod. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), pod)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create pod as it already exists", pod)
		} else {
			log.Error("Failed to create pod. ", err)
			return
		}
	}
	service := getService(namespace)
	err = controllerutil.SetControllerReference(operator, service, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for service. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), service)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create service as it already exists", service)
		} else {
			log.Error("Failed to create service. ", err)
			return
		}
	}
	route := getRoute(namespace)
	err = controllerutil.SetControllerReference(operator, route, reconciler.Service.GetScheme())
	if err != nil {
		log.Error("Failed to set owner reference for route. ", err)
		return
	}
	err = reconciler.Service.Create(context.TODO(), route)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Debug("Could not create route as it already exists", route)
		} else {
			log.Error("Failed to create route. ", err)
			return
		}
	}
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
	log.Debugf("Own image retrieved as %v", image)
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
				APIGroups: []string{v1.SchemeGroupVersion.Group},
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
