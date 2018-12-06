package defaults

import (
	"reflect"

	"github.com/imdario/mergo"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func merge(baseline v1.Environment, overwrite v1.Environment) (v1.Environment, error) {
	var env v1.Environment
	env.Console = mergeCustomObject(baseline.Console, overwrite.Console)
	if len(baseline.Others) == 0 {
		env.Others = overwrite.Others
	} else {
		for index := range baseline.Others {
			mergedObject := mergeCustomObject(baseline.Others[index], overwrite.Others[index])
			env.Others = append(env.Others, mergedObject)
		}
	}
	if len(baseline.Servers) != len(overwrite.Servers) {
		return v1.Environment{}, errors.New("Incompatible objects with different array lengths cannot be merged")
	}
	for index := range baseline.Servers {
		mergedObject := mergeCustomObject(baseline.Servers[index], overwrite.Servers[index])
		env.Servers = append(env.Servers, mergedObject)
	}
	return env, nil
}

func mergeCustomObject(baseline v1.CustomObject, overwrite v1.CustomObject) v1.CustomObject {
	var object v1.CustomObject
	object.PersistentVolumeClaims = mergePersistentVolumeClaims(baseline.PersistentVolumeClaims, overwrite.PersistentVolumeClaims)
	object.ServiceAccounts = mergeServiceAccounts(baseline.ServiceAccounts, overwrite.ServiceAccounts)
	object.Secrets = mergeSecrets(baseline.Secrets, overwrite.Secrets)
	object.Roles = mergeRoles(baseline.Roles, overwrite.Roles)
	object.RoleBindings = mergeRoleBindings(baseline.RoleBindings, overwrite.RoleBindings)
	object.DeploymentConfigs = mergeDeploymentConfigs(baseline.DeploymentConfigs, overwrite.DeploymentConfigs)
	object.Services = mergeServices(baseline.Services, overwrite.Services)
	object.Routes = mergeRoutes(baseline.Routes, overwrite.Routes)
	return object
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

func mergeRoles(baseline []rbacv1.Role, overwrite []rbacv1.Role) []rbacv1.Role {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		baselineRefs := getRoleReferenceSlice(baseline)
		overwriteRefs := getRoleReferenceSlice(overwrite)
		slice := make([]rbacv1.Role, combinedSize(baselineRefs, overwriteRefs))
		err := mergeObjects(baselineRefs, overwriteRefs, slice)
		if err != nil {
			logrus.Errorf("%v", err)
			return nil
		}
		return slice
	}
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

func getRoleReferenceSlice(objects []rbacv1.Role) []v1.OpenShiftObject {
	slice := make([]v1.OpenShiftObject, len(objects))
	for index := range objects {
		slice[index] = &objects[index]
	}
	return slice
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
	}
	if len(baseline) == 0 {
		return overwrite
	}
	baselineRefs := getDeploymentConfigReferenceSlice(baseline)
	overwriteRefs := getDeploymentConfigReferenceSlice(overwrite)
	for overwriteIndex := range overwrite {
		overwriteItem := &overwrite[overwriteIndex]
		baselineIndex, _ := findOpenShiftObject(overwriteItem, baselineRefs)
		if baselineIndex >= 0 {
			baselineItem := baseline[baselineIndex]
			err := mergo.Merge(&overwriteItem.ObjectMeta, baselineItem.ObjectMeta)
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

func mergeSpec(baseline appsv1.DeploymentConfigSpec, overwrite appsv1.DeploymentConfigSpec) (appsv1.DeploymentConfigSpec, error) {
	mergedTemplate, err := mergeTemplate(baseline.Template, overwrite.Template)
	if err != nil {
		return appsv1.DeploymentConfigSpec{}, err
	}
	overwrite.Template = mergedTemplate

	mergedTriggers, err := mergeTriggers(baseline.Triggers, overwrite.Triggers)
	if err != nil {
		return appsv1.DeploymentConfigSpec{}, err
	}
	overwrite.Triggers = mergedTriggers

	err = mergo.Merge(&baseline, overwrite, mergo.WithOverride)
	if err != nil {
		return appsv1.DeploymentConfigSpec{}, nil
	}
	return baseline, nil
}

func mergeTemplate(baseline *corev1.PodTemplateSpec, overwrite *corev1.PodTemplateSpec) (*corev1.PodTemplateSpec, error) {
	if overwrite == nil {
		return baseline, nil
	}
	err := mergo.Merge(&overwrite.ObjectMeta, baseline.ObjectMeta)
	if err != nil {
		logrus.Errorf("%v", err)
		return nil, nil
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

func mergeTriggers(baseline appsv1.DeploymentTriggerPolicies, overwrite appsv1.DeploymentTriggerPolicies) (appsv1.DeploymentTriggerPolicies, error) {
	var mergedTriggers []appsv1.DeploymentTriggerPolicy
	for baselineIndex, baselineItem := range baseline {
		_, found := findDeploymentTriggerPolicy(baselineItem, overwrite)
		if found == (appsv1.DeploymentTriggerPolicy{}) {
			logrus.Debugf("Not found, adding %v to slice\n", baselineItem)
		} else {
			logrus.Debugf("Will merge %v on top of %v\n", found, baselineItem)
			if baselineItem.ImageChangeParams != nil {
				if found.ImageChangeParams == nil {
					found.ImageChangeParams = baselineItem.ImageChangeParams
				} else {
					mergedImageChangeParams, err := mergeImageChangeParams(*baselineItem.ImageChangeParams, *found.ImageChangeParams)
					if err != nil {
						return nil, err
					}
					found.ImageChangeParams = &mergedImageChangeParams
				}
			}
			err := mergo.Merge(&baseline[baselineIndex], found, mergo.WithOverride)
			if err != nil {
				return nil, err
			}
		}
		mergedTriggers = append(mergedTriggers, baseline[baselineIndex])
	}
	for overwriteIndex, overwriteItem := range overwrite {
		_, found := findDeploymentTriggerPolicy(overwriteItem, mergedTriggers)
		if found == (appsv1.DeploymentTriggerPolicy{}) {
			logrus.Debugf("Not found, appending %v to slice\n", overwriteItem)
			mergedTriggers = append(mergedTriggers, overwrite[overwriteIndex])
		}
	}
	return mergedTriggers, nil
}

func mergeImageChangeParams(baseline appsv1.DeploymentTriggerImageChangeParams, overwrite appsv1.DeploymentTriggerImageChangeParams) (appsv1.DeploymentTriggerImageChangeParams, error) {
	err := mergo.Merge(&baseline, overwrite, mergo.WithOverride)
	if err != nil {
		return appsv1.DeploymentTriggerImageChangeParams{}, err
	}
	return baseline, nil
}

func findDeploymentTriggerPolicy(object appsv1.DeploymentTriggerPolicy, slice []appsv1.DeploymentTriggerPolicy) (int, appsv1.DeploymentTriggerPolicy) {
	for index, candidate := range slice {
		if candidate.Type == object.Type {
			return index, candidate
		}
	}
	return -1, appsv1.DeploymentTriggerPolicy{}
}

func mergePodSpecs(baseline corev1.PodSpec, overwrite corev1.PodSpec) (corev1.PodSpec, error) {
	mergedContainers, err := mergeContainers(baseline.Containers, overwrite.Containers)
	if err != nil {
		return corev1.PodSpec{}, err
	}
	overwrite.Containers = mergedContainers

	mergedVolumes, err := mergeVolumes(baseline.Volumes, overwrite.Volumes)
	if err != nil {
		return corev1.PodSpec{}, err
	}
	overwrite.Volumes = mergedVolumes

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
	if baseline[0].Env == nil {
		baseline[0].Env = make([]corev1.EnvVar, 0)
	}
	overwrite[0].Env = shared.EnvOverride(baseline[0].Env, overwrite[0].Env)
	mergedPorts, err := mergePorts(baseline[0].Ports, overwrite[0].Ports)
	if err != nil {
		return nil, err
	}
	overwrite[0].Ports = mergedPorts

	mergedVolumeMounts, err := mergeVolumeMounts(baseline[0].VolumeMounts, overwrite[0].VolumeMounts)
	if err != nil {
		return []corev1.Container{}, err
	}
	overwrite[0].VolumeMounts = mergedVolumeMounts

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

func mergeVolumes(baseline []corev1.Volume, overwrite []corev1.Volume) ([]corev1.Volume, error) {
	var mergedVolumes []corev1.Volume
	for baselineIndex, baselineItem := range baseline {
		idx, found := findVolume(baselineItem, overwrite)
		if idx == -1 {
			logrus.Debugf("Not found, adding %v to slice\n", baselineItem)
		} else {
			logrus.Debugf("Will merge %v on top of %v\n", found, baselineItem)
			err := mergo.Merge(&baseline[baselineIndex], found, mergo.WithOverride)
			if err != nil {
				return nil, err
			}
		}
		mergedVolumes = append(mergedVolumes, baseline[baselineIndex])
	}
	for overwriteIndex, overwriteItem := range overwrite {
		idx, _ := findVolume(overwriteItem, mergedVolumes)
		if idx == -1 {
			logrus.Debugf("Not found, appending %v to slice\n", overwriteItem)
			mergedVolumes = append(mergedVolumes, overwrite[overwriteIndex])
		}
	}
	return mergedVolumes, nil
}

func mergeVolumeMounts(baseline []corev1.VolumeMount, overwrite []corev1.VolumeMount) ([]corev1.VolumeMount, error) {
	var mergedVolumeMounts []corev1.VolumeMount
	for baselineIndex, baselineItem := range baseline {
		idx, found := findVolumeMount(baselineItem, overwrite)
		if idx == -1 {
			logrus.Debugf("Not found, adding %v to slice\n", baselineItem)
		} else {
			logrus.Debugf("Will merge %v on top of %v\n", found, baselineItem)
			err := mergo.Merge(&baseline[baselineIndex], found, mergo.WithOverride)
			if err != nil {
				return nil, err
			}
		}
		mergedVolumeMounts = append(mergedVolumeMounts, baseline[baselineIndex])
	}
	for overwriteIndex, overwriteItem := range overwrite {
		idx, _ := findVolumeMount(overwriteItem, mergedVolumeMounts)
		if idx == -1 {
			logrus.Debugf("Not found, appending %v to slice\n", overwriteItem)
			mergedVolumeMounts = append(mergedVolumeMounts, overwrite[overwriteIndex])
		}
	}
	return mergedVolumeMounts, nil
}

func findVolume(object corev1.Volume, slice []corev1.Volume) (int, corev1.Volume) {
	for index, candidate := range slice {
		if candidate.Name == object.Name {
			return index, candidate
		}
	}
	return -1, corev1.Volume{}
}

func findVolumeMount(object corev1.VolumeMount, slice []corev1.VolumeMount) (int, corev1.VolumeMount) {
	for index, candidate := range slice {
		if candidate.Name == object.Name {
			return index, candidate
		}
	}
	return -1, corev1.VolumeMount{}
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
		for overwriteIndex := range overwrite {
			overwriteItem := &overwrite[overwriteIndex]
			baselineIndex, _ := findOpenShiftObject(overwriteItem, baselineRefs)
			if baselineIndex >= 0 {
				baselineItem := baseline[baselineIndex]
				err := mergo.Merge(&overwriteItem.ObjectMeta, baselineItem.ObjectMeta)
				if err != nil {
					logrus.Errorf("%v", err)
					return nil
				}
				overwriteItem.Spec.Ports = mergeServicePorts(baselineItem.Spec.Ports, overwriteItem.Spec.Ports)
			}
		}
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

func mergeServicePorts(baseline []corev1.ServicePort, overwrite []corev1.ServicePort) []corev1.ServicePort {
	if len(overwrite) == 0 {
		return baseline
	} else if len(baseline) == 0 {
		return overwrite
	} else {
		var mergedServicePorts []corev1.ServicePort
		for baselineIndex, baselinePort := range baseline {
			found, servicePort := findServicePort(baselinePort, overwrite)
			if found {
				mergedServicePorts = append(mergedServicePorts, servicePort)
			} else {
				mergedServicePorts = append(mergedServicePorts, baseline[baselineIndex])
			}
		}
		for overwriteIndex, overwritePort := range overwrite {
			found, _ := findServicePort(overwritePort, baseline)
			if !found {
				mergedServicePorts = append(mergedServicePorts, overwrite[overwriteIndex])
			}
		}
		return mergedServicePorts
	}
}

func findServicePort(port corev1.ServicePort, ports []corev1.ServicePort) (bool, corev1.ServicePort) {
	for index, candidate := range ports {
		if port.Name == candidate.Name {
			return true, ports[index]
		}
	}
	return false, corev1.ServicePort{}
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
