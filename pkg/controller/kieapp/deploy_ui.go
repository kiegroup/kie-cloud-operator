package kieapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/components"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var consoleName = "console-cr-form"
var operatorName = os.Getenv(constants.OpNameEnv)
var caConfigMapName = operatorName + "-trusted-cabundle"

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
	role := getRole(namespace)
	roleBinding := getRoleBinding(namespace)
	sa := getServiceAccount(namespace)
	image := getImage(operator)
	pod := getPod(namespace, image, sa.Name, reconciler.OcpVersion, operator)
	service := getService(namespace, reconciler.OcpVersion)
	route := getRoute(namespace)
	scheme := reconciler.Service.GetScheme()
	// `inject-trusted-cabundle` ConfigMap only supported in OCP 4.2+
	if semver.Compare(reconciler.OcpVersion, "v4.2") >= 0 || reconciler.OcpVersion == "" {
		existing := &corev1.ConfigMap{}
		new := getCaConfigMap(namespace)
		controllerutil.SetOwnerReference(operator, new, scheme)
		if err := reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: new.Name, Namespace: new.Namespace}, existing); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Creating ConfigMap ", new.Name)
				if err := reconciler.Service.Create(context.TODO(), new); err != nil {
					log.Error("failed to create configmap", err)
				}
			} else {
				log.Error("failed to get configmap", err)
			}
		} else {
			if !reflect.DeepEqual(existing.Labels, new.Labels) {
				existing.Labels = new.Labels
				controllerutil.SetOwnerReference(operator, existing, scheme)
				log.Info("Updating ConfigMap ", existing.Name)
				if err := reconciler.Service.Update(context.TODO(), existing); err != nil {
					log.Error("failed to update configmap", err)
				}
			}
		}
	}
	requestedResources := []resource.KubernetesResource{role, roleBinding, sa, pod, service, route}
	resourceMap := compare.NewMapBuilder()
	for _, resource := range requestedResources {
		resourceMap.Add(resource)
	}
	deployed, err := loadCounterparts(reconciler, namespace, resourceMap.ResourceMap())
	if err != nil {
		log.Error("Failed to load deployed resources.", err)
		return
	}
	if _, err = reconciler.reconcileResources(operator, requestedResources, deployed); err != nil {
		log.Error("Failed to reconcile resources.", err)
		return
	}
	updateCSVlinks(reconciler, route, operator)
}

func loadCounterparts(reconciler *Reconciler, namespace string, requestedMap map[reflect.Type][]resource.KubernetesResource) (map[reflect.Type][]resource.KubernetesResource, error) {
	reader := read.New(reconciler.Service).WithNamespace(namespace)
	var deployedArray []resource.KubernetesResource
	for resourceType, requestedArray := range requestedMap {
		for _, requested := range requestedArray {
			deployed, err := reader.Load(resourceType, requested.GetName())
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
	}
	deployedMap := compare.NewMapBuilder().Add(deployedArray...).ResourceMap()
	return deployedMap, nil
}

func updateCSVlinks(reconciler *Reconciler, route *routev1.Route, operator *appsv1.Deployment) {
	var patch bool
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
	if reflect.ValueOf(csv).IsNil() {
		log.Info("No ClusterServiceVersion found, likely because no such owner was set on the operator. This might be because the operator was not installed through OLM")
		return
	}
	newCSV := csv.DeepCopy()
	link := getConsoleLink(newCSV)
	if link == nil || link.URL != url {
		patch = true
		if link == nil {
			newCSV.Spec.Links = append([]operatorsv1alpha1.AppLink{{Name: constants.ConsoleLinkName, URL: url}}, newCSV.Spec.Links...)
		} else {
			link.URL = url
		}
	}
	if !strings.Contains(csv.Spec.Description, constants.ConsoleDescription) {
		patch = true
		newCSV.Spec.Description = strings.Join([]string{newCSV.Spec.Description, constants.ConsoleDescription}, "\n\n")
	}
	if patch {
		log.Debugf("Patching ", csv.Name, " ", csv.Kind, ".")
		if err := patchCSVObject(reconciler, csv, newCSV); err != nil {
			log.Error("Failed to patch CSV. ", err)
		}
	}
}

func patchCSVObject(reconciler *Reconciler, cur, mod *operatorsv1alpha1.ClusterServiceVersion) error {
	patch, err := client.MergeFrom(cur).Data(mod)
	if err != nil || len(patch) == 0 || string(patch) == "{}" {
		return err
	}
	return reconciler.Service.Patch(context.TODO(), cur, client.RawPatch(types.MergePatchType, patch))
}

func getConsoleLink(csv *operatorsv1alpha1.ClusterServiceVersion) *operatorsv1alpha1.AppLink {
	for i, link := range csv.Spec.Links {
		if link.Name == constants.ConsoleLinkName {
			return &csv.Spec.Links[i]
		}
	}
	return nil
}

func getPod(namespace, image, sa, ocpVersion string, operator *appsv1.Deployment) *corev1.Pod {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
	}
	volume := corev1.Volume{Name: operatorName + "-proxy-tls"}
	volume.Secret = &corev1.SecretVolumeSource{SecretName: volume.Name}
	sar, err := json.Marshal(map[string]string{
		"name":      consoleName,
		"namespace": namespace,
		"resource":  "services",
		"verb":      "patch",
	})
	if err != nil {
		log.Error("Failed to marshal sar config to json. ", err)
	}
	debug := constants.DebugFalse
	if shared.EnvVarSet(constants.DebugTrue, operator.Spec.Template.Spec.Containers[0].Env) {
		debug = constants.DebugTrue
	}
	var ocpMajor, ocpMinor string
	splitVersion := strings.Split(strings.TrimPrefix(ocpVersion, "v"), ".")
	if len(splitVersion) > 1 {
		ocpMajor = splitVersion[0]
		ocpMinor = splitVersion[1]
	}
	// set oauth image by ocp version, default to latest available
	oauthImage := constants.Oauth4ImageLatestURL
	if val, exists := os.LookupEnv(fmt.Sprintf(constants.OauthVar+"%s.%s", ocpMajor, ocpMinor)); exists {
		oauthImage = val
	} else if val, exists := os.LookupEnv(constants.OauthVar + "LATEST"); exists {
		oauthImage = val
	}
	sarString := fmt.Sprintf("--openshift-sar=%s", sar)
	httpPort := int32(8080)
	httpsPort := int32(8443)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: sa,
			Volumes:            []corev1.Volume{volume},
			Containers: []corev1.Container{
				{
					Name:            "oauth-proxy",
					Image:           oauthImage,
					ImagePullPolicy: operator.Spec.Template.Spec.Containers[0].ImagePullPolicy,
					Ports:           []corev1.ContainerPort{{Name: "public", ContainerPort: httpsPort}},
					Args: []string{
						"--http-address=",
						fmt.Sprintf("--https-address=:%d", httpsPort),
						fmt.Sprintf("--upstream=http://localhost:%d", httpPort),
						"--provider=openshift",
						sarString,
						fmt.Sprintf("--openshift-service-account=%s", sa),
						"--openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
						"--tls-cert=/etc/tls/private/tls.crt",
						"--tls-key=/etc/tls/private/tls.key",
						"--cookie-secret=SECRET",
					},
					VolumeMounts: []corev1.VolumeMount{{Name: volume.Name, MountPath: "/etc/tls/private"}},
				},
				{
					Name:            consoleName,
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
	// `inject-trusted-cabundle` ConfigMap only supported in OCP 4.2+
	if semver.Compare(ocpVersion, "v4.2") >= 0 || ocpVersion == "" {
		caVolume := corev1.Volume{
			Name: caConfigMapName,
		}
		caVolume.ConfigMap = &corev1.ConfigMapVolumeSource{
			Items: []corev1.KeyToPath{{Key: "ca-bundle.crt", Path: "ca-bundle.crt"}},
		}
		caVolume.ConfigMap.Name = caVolume.Name
		pod.Spec.Volumes = append(pod.Spec.Volumes, caVolume)
		mountPath := "/etc/pki/ca-trust/extracted/crt"
		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, "--openshift-ca="+mountPath+"/"+caVolume.ConfigMap.Items[0].Key)
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: caVolume.Name, MountPath: mountPath, ReadOnly: true})
	} else {
		log.Warn(err)
	}
	return pod
}

func getImage(operator *appsv1.Deployment) string {
	image := operator.Spec.Template.Spec.Containers[0].Image
	return image
}

func getService(namespace, ocpVersion string) *corev1.Service {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consoleName,
			Namespace:   namespace,
			Annotations: map[string]string{"service.beta.openshift.io/serving-cert-secret-name": operatorName + "-proxy-tls"},
			Labels:      labels,
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
	return svc
}

func getRoute(namespace string) *routev1.Route {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: consoleName,
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationReencrypt,
			},
		},
	}
}

func getCaConfigMap(namespace string) *corev1.ConfigMap {
	labels := map[string]string{
		"app": operatorName,
		"config.openshift.io/inject-trusted-cabundle": "true",
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caConfigMapName,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func getRole(namespace string) *rbacv1.Role {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
	}
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleName,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{api.SchemeGroupVersion.Group},
				Resources: []string{"kieapps"},
				Verbs:     components.Verbs,
			},
		},
	}
}

func getRoleBinding(namespace string) *rbacv1.RoleBinding {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
	}
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleName,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: consoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: consoleName,
			},
		},
	}
}

func getServiceAccount(namespace string) *corev1.ServiceAccount {
	labels := map[string]string{
		"app":  operatorName,
		"name": consoleName,
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
			Name: consoleName,
		},
	})
	if err != nil {
		log.Error("Failed to marshal annotation config to json. ", err)
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consoleName,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.primary": string(annotation),
			},
		},
	}
}
