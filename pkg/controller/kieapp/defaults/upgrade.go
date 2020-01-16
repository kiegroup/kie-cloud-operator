package defaults

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/google/go-cmp/cmp"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// checkProductUpgrade ...
func checkProductUpgrade(cr *api.KieApp) (minor, micro bool, err error) {
	setDefaults(cr)
	if checkVersion(cr.Spec.Version) {
		if cr.Spec.Version != constants.CurrentVersion && cr.Spec.Upgrades.Enabled {
			micro = cr.Spec.Upgrades.Enabled
			minor = cr.Spec.Upgrades.Minor
		}
	} else {
		err = fmt.Errorf("Product version %s is not allowed in operator version %s. The following versions are allowed - %s", cr.Spec.Version, version.Version, constants.SupportedVersions)
	}
	return minor, micro, err
}

// checkVersion ...
func checkVersion(productVersion string) bool {
	for _, version := range constants.SupportedVersions {
		if version == productVersion {
			return true
		}
	}
	return false
}

// GetMinorImageVersion ...
func GetMinorImageVersion(productVersion string) string {
	major, minor, _ := MajorMinorMicro(productVersion)
	return major + minor
}

// MajorMinorMicro ...
func MajorMinorMicro(productVersion string) (major, minor, micro string) {
	version := strings.Split(productVersion, ".")
	for len(version) < 3 {
		version = append(version, "0")
	}
	return version[0], version[1], version[2]
}

// getConfigVersionDiffs ...
func getConfigVersionDiffs(fromVersion, toVersion string, service api.PlatformService) error {
	if checkVersion(fromVersion) && checkVersion(toVersion) {
		fromList, toList := getConfigVersionLists(fromVersion, toVersion)
		diffs := configDiffs(fromList, toList)
		cmDiffs := diffs
		// only check against existing configmaps if running via deployment in a cluster
		if _, depNameSpace, useEmbedded := UseEmbeddedFiles(service); !useEmbedded {
			cmFromList := map[string][]map[string]string{}
			for name := range fromList {
				nameSplit := strings.Split(name, "-")
				cmName := strings.Join(append([]string{nameSplit[0], fromVersion}, nameSplit[1:]...), "-")
				// *** need to retrieve cm of same name w/ current version and do compare against default upgrade diffs...
				currentCM := &corev1.ConfigMap{}
				if err := service.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: depNameSpace}, currentCM); err != nil {
					return err
				}
				cmFromList[name] = append(cmFromList[name], currentCM.Data)
			}
			cmDiffs = configDiffs(cmFromList, toList)
		} else if service.IsMockService() { // test
			fromList[constants.ConfigMapPrefix] = []map[string]string{{"common.yaml": "changed"}}
			cmDiffs = configDiffs(fromList, toList)
		}
		// if conflicts, stop upgrade
		// COMPARE NEEDS IMPROVEMENT - more precise comparison? and should maybe show exact differences that conflict.
		if !cmp.Equal(diffs, cmDiffs) {
			return fmt.Errorf("Can't upgrade, potential configuration conflicts in your %s ConfigMap(s)", fromVersion)
		}
	}
	return nil
}

// getConfigVersionLists ...
func getConfigVersionLists(fromVersion, toVersion string) (configFromList, configToList map[string][]map[string]string) {
	fromList := map[string][]map[string]string{}
	toList := map[string][]map[string]string{}
	if checkVersion(fromVersion) && checkVersion(toVersion) {
		box := packr.New("config", "../../../../config")
		if box.HasDir(fromVersion) && box.HasDir(toVersion) {
			cmList := getCMListfromBox(box)
			for cmName, cmData := range cmList {
				cmSplit := strings.Split(cmName, "-")
				name := strings.Join(append(cmSplit[:1], cmSplit[2:]...), "-")
				if cmSplit[1] == fromVersion {
					fromList[name] = cmData
				}
				if cmSplit[1] == toVersion {
					toList[name] = cmData
				}
			}
		}
	}
	return fromList, toList
}

// configDiffs ...
func configDiffs(cmFromList, cmToList map[string][]map[string]string) map[string]string {
	configDiffs := map[string]string{}
	for cmName, fromMapSlice := range cmFromList {
		if toMapSlice, ok := cmToList[cmName]; ok {
			diff := cmp.Diff(fromMapSlice, toMapSlice)
			if diff != "" {
				configDiffs[cmName] = diff
			}
		}
	}
	return configDiffs
}
