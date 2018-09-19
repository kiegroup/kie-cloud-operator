package kieserver

import (
	"github.com/kiegroup/kie-cloud-operator/internal/constants"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/defaults"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/rhpam/v1alpha1"
	"github.com/openshift/api/apps"
	"github.com/openshift/api/apps/v1"
	"github.com/openshift/api/route"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetKieServer(cr *v1alpha1.App) []runtime.Object {
	_, serviceName, labels := shared.GetCommonLabels(cr, constants.KieServerServicePrefix)
	image := shared.GetImage(cr.Spec.Server.Image, "rhpam70-kieserver-openshift")

	dc := v1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: apps.GroupName + "/v1", //TODO find out if there is a function that provides this
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "App",
				}),
			},
			Labels: labels,
		},
		Spec: v1.DeploymentConfigSpec{
			Strategy: v1.DeploymentStrategy{
				Type: v1.DeploymentStrategyTypeRecreate,
			},
			Triggers: v1.DeploymentTriggerPolicies{
				{
					Type: v1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &v1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{serviceName},
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Namespace: constants.ImageStreamNamespace,
							Name:      constants.KieServerImageStreamName + ":" + constants.ImageStreamTag,
						},
					},
				},
			},
			Replicas: 1,
			Selector: labels,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &[]int64{60}[0],
					Containers: []corev1.Container{
						{
							Name:            serviceName,
							Image:           image,
							ImagePullPolicy: "Always",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("220Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("220Mi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/bash", "-c", "curl --fail --silent -u 'adminUser:RedHat' http://localhost:8080/services/rest/server/healthcheck"},
									},
								},
								InitialDelaySeconds: 180,
								TimeoutSeconds:      2,
								PeriodSeconds:       15,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/bash", "-c", "curl --fail --silent -u 'adminUser:RedHat' http://localhost:8080/services/rest/server/readycheck"},
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      2,
								PeriodSeconds:       30,
								FailureThreshold:    6,
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "jolokia",
									ContainerPort: 8778,
									Protocol:      "TCP",
								},
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}
	rhpamcentrServiceName := cr.ObjectMeta.Name + "-" + constants.RhpamcentrServicePrefix
	defaultEnv := defaults.ServerEnvironmentDefaults()
	defaultEnv["KIE_SERVER_CONTROLLER_SERVICE"] = rhpamcentrServiceName
	defaultEnv["RHPAMCENTR_MAVEN_REPO_SERVICE"] = rhpamcentrServiceName
	shared.MergeContainerConfigs(dc.Spec.Template.Spec.Containers, cr.Spec.Server, defaultEnv)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "App",
				}),
			},
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{"deploymentconfig": serviceName},
		},
	}

	openshiftRoute := routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: route.GroupName + "/v1", //TODO find out if there is a function that provides this
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "App",
				}),
			},
			Labels: labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Name: serviceName,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
		},
	}
	//logrus.Infof("Will skip over creating %s" + deployment.Kind)
	return []runtime.Object{dc.DeepCopyObject(), service, openshiftRoute.DeepCopyObject()}
}
