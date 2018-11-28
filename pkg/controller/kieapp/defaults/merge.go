package defaults

import (
	"github.com/imdario/mergo"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func merge(baseline *v1.CustomObject, overwrite *v1.CustomObject) {
	baseline.PersistentVolumeClaims = mergePersistentVolumeClaims(baseline.PersistentVolumeClaims, overwrite.PersistentVolumeClaims)
	baseline.ServiceAccounts = mergeServiceAccounts(baseline.ServiceAccounts, overwrite.ServiceAccounts)
	baseline.Secrets = mergeSecrets(baseline.Secrets, overwrite.Secrets)
	baseline.RoleBindings = mergeRoleBindings(baseline.RoleBindings, overwrite.RoleBindings)
	baseline.DeploymentConfigs = mergeDeploymentConfigs(baseline.DeploymentConfigs, overwrite.DeploymentConfigs)
	baseline.Services = mergeServices(baseline.Services, overwrite.Services)
	baseline.Routes = mergeRoutes(baseline.Routes, overwrite.Routes)
}

func mergePersistentVolumeClaims(baseline []corev1.PersistentVolumeClaim, overwrite []corev1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getPersistentVolumeClaimReferenceSlice(baseline)
		overwriteRefs := getPersistentVolumeClaimReferenceSlice(overwrite)
		slice := make([]corev1.PersistentVolumeClaim, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func getPersistentVolumeClaimReferenceSlice(objects []corev1.PersistentVolumeClaim) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeServiceAccounts(baseline []corev1.ServiceAccount, overwrite []corev1.ServiceAccount) []corev1.ServiceAccount {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getServiceAccountReferenceSlice(baseline)
		overwriteRefs := getServiceAccountReferenceSlice(overwrite)
		slice := make([]corev1.ServiceAccount, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func getServiceAccountReferenceSlice(objects []corev1.ServiceAccount) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeSecrets(baseline []corev1.Secret, overwrite []corev1.Secret) []corev1.Secret {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getSecretReferenceSlice(baseline)
		overwriteRefs := getSecretReferenceSlice(overwrite)
		slice := make([]corev1.Secret, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func getSecretReferenceSlice(objects []corev1.Secret) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeRoleBindings(baseline []rbacv1.RoleBinding, overwrite []rbacv1.RoleBinding) []rbacv1.RoleBinding {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getRoleBindingReferenceSlice(baseline)
		overwriteRefs := getRoleBindingReferenceSlice(overwrite)
		slice := make([]rbacv1.RoleBinding, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func getRoleBindingReferenceSlice(objects []rbacv1.RoleBinding) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeDeploymentConfigs(baseline []appsv1.DeploymentConfig, overwrite []appsv1.DeploymentConfig) []appsv1.DeploymentConfig {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getDeploymentConfigReferenceSlice(baseline)
		overwriteRefs := getDeploymentConfigReferenceSlice(overwrite)
		for overwriteIndex := range overwrite {
			overwriteItem := overwrite[overwriteIndex]
			baselineIndex, _ := findOpenShiftObject(&overwriteItem, baselineRefs)
			if baselineIndex >= 0 {
				baselineItem := baseline[baselineIndex]
				err := mergeLabels(&overwriteItem.ObjectMeta, &baselineItem.ObjectMeta) //reverse merge to maintain changes
				if err != nil {
					logrus.Errorf("%v", err)
					return nil
				}
				err = mergo.Merge(&overwriteItem.ObjectMeta, baselineItem.ObjectMeta)
				if err != nil {
					logrus.Errorf("%v", err)
					return nil
				}
				mergedSpec, err := mergeSpec(baselineItem.Spec, overwriteItem.Spec)
				if err != nil {
					logrus.Errorf("%v", err)
					return nil
				}
				overwriteItem.Spec = mergedSpec
			}
		}
		slice := make([]appsv1.DeploymentConfig, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func mergeLabels(baseline metav1.Object, overwrite metav1.Object) error {
	mergedLabels := baseline.GetLabels()
	err := mergo.Merge(&mergedLabels, overwrite.GetLabels(), mergo.WithOverride)
	if err != nil {
		return err
	}
	baseline.SetLabels(mergedLabels)
	return nil
}

func mergeSpec(baseline appsv1.DeploymentConfigSpec, overwrite appsv1.DeploymentConfigSpec) (appsv1.DeploymentConfigSpec, error) {
	mergedTemplate, err := mergeTemplate(baseline.Template, overwrite.Template)
	if err != nil {
		return appsv1.DeploymentConfigSpec{}, err
	}
	overwrite.Template = mergedTemplate

	err = mergo.Merge(&baseline, overwrite, mergo.WithOverride)
	if err != nil {
		return appsv1.DeploymentConfigSpec{}, nil
	}
	return baseline, nil
}

func mergeTemplate(baseline *corev1.PodTemplateSpec, overwrite *corev1.PodTemplateSpec) (*corev1.PodTemplateSpec, error) {
	err := mergeLabels(overwrite, baseline)
	if err != nil {
		return nil, err
	}
	mergedPodSpec, err := mergePodSpecs(baseline.Spec, overwrite.Spec)
	if err != nil {
		return nil, err
	}
	overwrite.Spec = mergedPodSpec

	err = mergo.Merge(baseline, *overwrite, mergo.WithOverride)
	if err != nil {
		return nil, err
	}
	return baseline, nil
}

func mergePodSpecs(baseline corev1.PodSpec, overwrite corev1.PodSpec) (corev1.PodSpec, error) {
	mergedContainers, err := mergeContainers(baseline.Containers, overwrite.Containers)
	if err != nil {
		return corev1.PodSpec{}, err
	}
	overwrite.Containers = mergedContainers

	err = mergo.Merge(&baseline, overwrite, mergo.WithOverride)
	if err != nil {
		return corev1.PodSpec{}, err
	}
	return baseline, nil
}

func mergeContainers(baseline []corev1.Container, overwrite []corev1.Container) ([]corev1.Container, error) {
	if len(overwrite) == 0 {
		return baseline, nil
	} else if len(baseline) == 0 {
		return overwrite, nil
	} else if len(baseline) > 1 || len(overwrite) > 1 {
		err := errors.New("Merge algorithm does not yet support multiple containers within a deployment")
		return nil, err
	}
	overwrite[0].Env = shared.EnvOverride(baseline[0].Env, overwrite[0].Env)
	mergedPorts, err := mergePorts(baseline[0].Ports, overwrite[0].Ports)
	if err != nil {
		return nil, err
	}
	overwrite[0].Ports = mergedPorts

	err = mergo.Merge(&baseline[0], overwrite[0], mergo.WithOverride)
	if err != nil {
		return nil, err
	}
	return baseline, nil
}

func mergePorts(baseline []corev1.ContainerPort, overwrite []corev1.ContainerPort) ([]corev1.ContainerPort, error) {
	var slice []corev1.ContainerPort
	for index := range baseline {
		found := findContainerPort(baseline[index], overwrite)
		if found != (corev1.ContainerPort{}) {
			err := mergo.Merge(&baseline[index], found, mergo.WithOverride)
			if err != nil {
				return nil, err
			}
		}
		slice = append(slice, baseline[index])
	}
	for index := range overwrite {
		found := findContainerPort(overwrite[index], baseline)
		if found == (corev1.ContainerPort{}) {
			slice = append(slice, overwrite[index])
		}
	}
	return slice, nil
}

func findContainerPort(port corev1.ContainerPort, ports []corev1.ContainerPort) corev1.ContainerPort {
	for index := range ports {
		if port.Name == ports[index].Name {
			return ports[index]
		}
	}
	return corev1.ContainerPort{}
}

func getDeploymentConfigReferenceSlice(objects []appsv1.DeploymentConfig) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeServices(baseline []corev1.Service, overwrite []corev1.Service) []corev1.Service {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getServiceReferenceSlice(baseline)
		overwriteRefs := getServiceReferenceSlice(overwrite)
		slice := make([]corev1.Service, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}
func getServiceReferenceSlice(objects []corev1.Service) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func mergeRoutes(baseline []routev1.Route, overwrite []routev1.Route) []routev1.Route {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getRouteReferenceSlice(baseline)
		overwriteRefs := getRouteReferenceSlice(overwrite)
		slice := make([]routev1.Route, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
}

func getRouteReferenceSlice(objects []routev1.Route) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
}

func combinedSize(baseline []v1.OpenShiftObject, overwrite []v1.OpenShiftObject) int {
	count := 0
	for _, object := range overwrite {
		_, found := findOpenShiftObject(object, baseline)
		if found == nil && object.GetAnnotations()["delete"] != "true" {
			//unique item with no counterpart in baseline, count it
			count++
		} else if found != nil && object.GetAnnotations()["delete"] == "true" {
			///Deletes the counterpart in baseline, deduct 1 since the counterpart is being counted below
			count--
		}
	}
	count += len(baseline)
	return count
}

func mergeObjects(baseline []v1.OpenShiftObject, overwrite []v1.OpenShiftObject, objectSlice interface{}) error {
	slice := reflect.ValueOf(objectSlice)
	sliceIndex := 0
	for _, object := range baseline {
		_, found := findOpenShiftObject(object, overwrite)
		if found == nil {
			slice.Index(sliceIndex).Set(reflect.ValueOf(object).Elem())
			sliceIndex++
			logrus.Debugf("Not found, added %s to beginning of slice\n", object)
		} else if found.GetAnnotations()["delete"] != "true" {
			err := mergo.Merge(object, found, mergo.WithOverride)
			if err != nil {
				return err
			}
			slice.Index(sliceIndex).Set(reflect.ValueOf(object).Elem())
			sliceIndex++
			if found.GetAnnotations() == nil {
				annotations := make(map[string]string)
				found.SetAnnotations(annotations)
			}
		}
	}
	for _, object := range overwrite {
		if object.GetAnnotations()["delete"] != "true" {
			_, found := findOpenShiftObject(object, baseline)
			if found == nil {
				slice.Index(sliceIndex).Set(reflect.ValueOf(object).Elem())
				sliceIndex++
			}
		}
	}
	return nil
}

func findOpenShiftObject(object v1.OpenShiftObject, slice []v1.OpenShiftObject) (int, v1.OpenShiftObject) {
	for index, candidate := range slice {
		if candidate.GetName() == object.GetName() {
			return index, candidate
		}
	}
	return -1, nil
}
