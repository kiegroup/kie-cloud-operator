package defaults

import (
	"fmt"
	"strings"

	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/version"
)

// checkProductUpgrade ...
func checkProductUpgrade(cr *api.KieApp) (minor, micro bool, err error) {
	if checkVersion(cr) {
		if cr.Spec.Version != constants.CurrentVersion && cr.Spec.Upgrades.Enabled {
			micro = cr.Spec.Upgrades.Enabled
			minor = cr.Spec.Upgrades.Minor
		}
	} else {
		err = fmt.Errorf("Product version %s is not supported in operator version %s. The following versions are supported - %s", cr.Spec.Version, version.Version, constants.SupportedVersions)
	}
	return minor, micro, err
}

// checkVersion ...
func checkVersion(cr *api.KieApp) bool {
	if cr.Spec.Version == "74" {
		cr.Spec.Version = "7.4.1"
	}
	for _, version := range constants.SupportedVersions {
		if version == cr.Spec.Version {
			return true
		}
	}
	return false
}

// getMinorImageVersion ...
func getMinorImageVersion(productVersion string) string {
	major, minor, _ := MajorMinorMicro(productVersion)
	return fmt.Sprintf("%s%s", major, minor)
}

// MajorMinorMicro ...
func MajorMinorMicro(productVersion string) (major, minor, micro string) {
	version := strings.Split(productVersion, ".")
	for len(version) < 3 {
		version = append(version, "0")
	}
	return version[0], version[1], version[2]
}
