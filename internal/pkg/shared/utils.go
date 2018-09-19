package shared

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/bmozaffa/rhpam-operator/pkg/apis/rhpam/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func GetCommonLabels(app *v1alpha1.App, service string) (string, string, map[string]string) {
	appName := app.ObjectMeta.Name
	serviceName := appName + "-" + service
	labels := map[string]string{
		"application": appName,
		"deployment":  serviceName,
		"service":     serviceName,
	}
	return appName, serviceName, labels
}

func GetImage(configuredString string, defaultString string) string {
	if len(configuredString) > 0 {
		return configuredString
	} else {
		return defaultString
	}
}

func GetResources(configuration corev1.ResourceRequirements) corev1.ResourceRequirements {

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: getQuantity(configuration.Limits.Memory(), "220Mi"),
		},
	}
}

func getQuantity(configuredQuantity *resource.Quantity, defaultQuantity string) resource.Quantity {
	if configuredQuantity.IsZero() {
		return resource.MustParse(defaultQuantity)
	} else {
		return *configuredQuantity
	}
}

func GetEnvVars(defaults map[string]string, vars []corev1.EnvVar) []corev1.EnvVar {
	for _, envVar := range vars {
		defaults[envVar.Name] = envVar.Value
	}
	allVars  := make([]corev1.EnvVar, len(defaults))
	index := 0
	for key, value := range defaults {
		allVars[index] = corev1.EnvVar{Name: key, Value: value}
		index++
	}
	return allVars
}
