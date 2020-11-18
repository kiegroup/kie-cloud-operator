package components

import (
	"sort"
	"strings"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	consolev1 "github.com/openshift/api/console/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	csvv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
		for _, i := range constants.Images {
			if i.Var == constants.PamProcessMigrationVar && semver.Compare(semver.MajorMinor("v"+imageVersion), "v7.8") < 0 {
				continue
			}
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  i.Var + imageVersion,
				Value: i.Registry + ":" + imageVersion,
			})
		}
		if versionConstants, found := constants.VersionConstants[imageVersion]; found {
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  constants.OseCliVar + imageVersion,
				Value: versionConstants.OseCliImageURL,
			})
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  constants.MySQLVar + imageVersion,
				Value: versionConstants.MySQLImageURL,
			})
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  constants.PostgreSQLVar + imageVersion,
				Value: versionConstants.PostgreSQLImageURL,
			})
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  constants.DatagridVar + imageVersion,
				Value: versionConstants.DatagridImageURL,
			})
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  constants.BrokerVar + imageVersion,
				Value: versionConstants.BrokerImageURL,
			})
		}
	}
	// add oauth-proxy image references
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  constants.OauthVar + "LATEST",
		Value: constants.Oauth4ImageLatestURL,
	})
	sort.Sort(sort.Reverse(sort.StringSlice(constants.Ocp4Versions)))
	for _, ocpVersion := range constants.Ocp4Versions {
		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  constants.OauthVar + ocpVersion,
			Value: constants.Oauth4ImageURL + ":v" + ocpVersion,
		})
	}
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  constants.OauthVar + "3",
		Value: constants.Oauth3ImageLatestURL,
	})

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
				},
				Resources: []string{
					"configmaps",
					"pods",
					"services",
					"services/finalizers",
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
					"images",
					"imagestreams",
					"imagestreamimages",
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
					"kieapps/status",
					"kieapps/finalizers",
				},
				Verbs: Verbs,
			},
			{
				APIGroups: []string{
					monv1.SchemeGroupVersion.Group,
				},
				Resources: []string{
					"servicemonitors",
				},
				Verbs: []string{
					"get",
					"create",
				},
			},
			{
				APIGroups: []string{
					csvv1.SchemeGroupVersion.Group,
				},
				Resources: []string{
					"clusterserviceversions",
					"subscriptions",
				},
				Verbs: []string{
					"get",
					"list",
					"patch",
					"update",
					"watch",
				},
			},
		},
	}
	return role
}

func GetClusterRole(operatorName string) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: operatorName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{consolev1.GroupVersion.Group},
				Resources: []string{"consolelinks", "consoleyamlsamples"},
				Verbs: []string{
					"get",
					"create",
					"update",
					"delete",
				},
			},
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
