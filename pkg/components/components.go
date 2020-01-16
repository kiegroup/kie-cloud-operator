package components

import (
	"sort"
	"strings"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var Verbs = []string{
	"create",
	"delete",
	"deletecollection",
	"get",
	"list",
	"patch",
	"update",
	"watch",
}

func GetDeployment(operatorName, repository, context, imageName, tag, imagePullPolicy string) *appsv1.Deployment {
	registryName := strings.Join([]string{repository, context, imageName}, "/")
	image := strings.Join([]string{registryName, tag}, ":")
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: operatorName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": operatorName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": operatorName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: operatorName,
					Containers: []corev1.Container{
						{
							Name:            operatorName,
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Command:         []string{"kie-cloud-operator"},
							Env: []corev1.EnvVar{
								{
									Name: "OPERATOR_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.labels['name']",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "OPERATOR_UI",
									Value: "true",
								},
								{
									Name:  "DEBUG",
									Value: "false",
								},
							},
						},
					},
				},
			},
		},
	}
	sort.Sort(sort.Reverse(sort.StringSlice(constants.SupportedVersions)))
	for _, imageVersion := range constants.SupportedVersions {
		if defaults.GetMinorImageVersion(imageVersion) >= "77" {
			for _, i := range constants.Images {
				env := corev1.EnvVar{
					Name:  i.Var + imageVersion,
					Value: i.Registry + ":" + imageVersion,
				}
				deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, env)
			}
		}
	}

	return deployment
}

func GetRole(operatorName string) *rbacv1.Role {
	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: operatorName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
					appsv1.SchemeGroupVersion.Group,
					oappsv1.SchemeGroupVersion.Group,
					rbacv1.SchemeGroupVersion.Group,
					routev1.SchemeGroupVersion.Group,
					buildv1.SchemeGroupVersion.Group,
					oimagev1.SchemeGroupVersion.Group,
					api.SchemeGroupVersion.Group,
				},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{
					monv1.SchemeGroupVersion.Group,
				},
				Resources: []string{"servicemonitors"},
				Verbs:     []string{"get", "create"},
			},
			{
				APIGroups: []string{
					csvv1.SchemeGroupVersion.Group,
				},
				Resources: []string{"clusterserviceversions"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{
					appsv1.SchemeGroupVersion.Group,
				},
				ResourceNames: []string{operatorName},
				Resources:     []string{"deployments/finalizers"},
				Verbs:         []string{"update"},
			},
			/*
				{
					APIGroups: []string{
						"",
					},
					Resources: []string{
						"configmaps",
						"pods",
						"services",
						"serviceaccounts",
						"persistentvolumeclaims",
						"secrets",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						appsv1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"deployments",
						"deployments/finalizers",
						"replicasets",
						"statefulsets",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						oappsv1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"deploymentconfigs",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						rbacv1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"rolebindings",
						"roles",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						routev1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"routes",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						buildv1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"buildconfigs",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						oimagev1.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"imagestreams",
						"imagestreamtags",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						api.SchemeGroupVersion.Group,
					},
					Resources: []string{
						"kieapps",
						"kieapps/finalizers",
					},
					Verbs: Verbs,
				},
				{
					APIGroups: []string{
						monv1.SchemeGroupVersion.Group,
					},
					Resources: []string{"servicemonitors"},
					Verbs:     []string{"get", "create"},
				},
				{
					APIGroups: []string{
						csvv1.SchemeGroupVersion.Group,
					},
					Resources: []string{"clusterserviceversions"},
					Verbs: []string{
						"get",
						"list",
						"patch",
						"update",
						"watch",
					},
				},
			*/
		},
	}
	return role
}

func GetCrd() *extv1beta1.CustomResourceDefinition {
	plural := "kieapps"
	crd := &extv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: extv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: plural + "." + api.SchemeGroupVersion.Group,
		},
		Spec: extv1beta1.CustomResourceDefinitionSpec{
			Scope:   "Namespaced",
			Group:   api.SchemeGroupVersion.Group,
			Version: api.SchemeGroupVersion.Version,
			Versions: []extv1beta1.CustomResourceDefinitionVersion{
				{
					Name:    api.SchemeGroupVersion.Version,
					Served:  true,
					Storage: true,
					Schema:  &extv1beta1.CustomResourceValidation{OpenAPIV3Schema: &extv1beta1.JSONSchemaProps{}},
				},
				{
					Name:    v1.SchemeGroupVersion.Version,
					Served:  true,
					Storage: false,
					Schema:  &extv1beta1.CustomResourceValidation{OpenAPIV3Schema: &extv1beta1.JSONSchemaProps{}},
				},
			},
			Names: extv1beta1.CustomResourceDefinitionNames{
				Plural:   "kieapps",
				ListKind: "KieAppList",
				Singular: "kieapp",
				Kind:     "KieApp",
			},
			AdditionalPrinterColumns: []extv1beta1.CustomResourceColumnDefinition{
				{Name: "Age", Type: "date", JSONPath: ".metadata.creationTimestamp"},
				{Name: "Phase", Type: "string", JSONPath: ".status.phase"},
			},
			Subresources: &extv1beta1.CustomResourceSubresources{
				Status: &extv1beta1.CustomResourceSubresourceStatus{},
			},
		},
	}
	return crd
}

func GetCR() *api.KieApp {
	return &api.KieApp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       "KieApp",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: api.KieAppSpec{
			Environment: api.RhpamTrial,
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
