package shared

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/rhpam/v1alpha1"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func GetCommonLabels(app *v1alpha1.App, service string) (string, string, map[string]string) {
	appName := app.ObjectMeta.Name
	serviceName := appName + "-" + service
	labels := map[string]string{
		"app":     appName,
		"service": serviceName,
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

func getEnvVars(defaults map[string]string, vars []corev1.EnvVar) []corev1.EnvVar {
	for _, envVar := range vars {
		defaults[envVar.Name] = envVar.Value
	}
	allVars := make([]corev1.EnvVar, len(defaults))
	index := 0
	for key, value := range defaults {
		allVars[index] = corev1.EnvVar{Name: key, Value: value}
		index++
	}
	return allVars
}

func MergeContainerConfigs(containers []corev1.Container, crc corev1.Container, defaultEnv map[string]string) []corev1.Container {
	crc.Env = getEnvVars(defaultEnv, crc.Env)
	/*
		unstructObj, err := k8sutil.UnstructuredFromRuntimeObject(object)
		if err != nil {
			return err
		}

		// Update the arg object with the result
		err = k8sutil.UnstructuredIntoRuntimeObject(unstructObj, object)
		if err != nil {
			return fmt.Errorf("failed to unmarshal the retrieved data: %v", err)
		}
	*/

	for i, c := range containers {
		/*
			patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, ct)
			if err != nil {
				logrus.Warnf("Failed to get merge info: %v", err)
			}
			_, err = strategicpatch.StrategicMergePatch(oldData, patch, ct)
			if err != nil {
				logrus.Warnf("Failed to merge container configs: %v", err)
			}
			err = json.Unmarshal(crcb, &ct)
			if err != nil {
				logrus.Warnf("Failed to unmarshal container configs: %v", err)
			}
		*/
		ct := c
		err := mergo.Merge(&ct, crc, mergo.WithOverride)
		if err != nil {
			logrus.Warnf("Failed to unmarshal container configs: %v", err)
		}
		containers[i] = ct
	}

	return containers
}
