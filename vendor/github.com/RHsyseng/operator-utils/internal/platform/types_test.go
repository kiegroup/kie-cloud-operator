package platform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestK8SVersionHelpers(t *testing.T) {

	ocpTestVersions := []struct {
		version string
		major   string
		minor   string
	}{
		{"1.11+", "1", "11+"},
		{"1.13+", "1", "13+"},
	}

	for _, v := range ocpTestVersions {

		info := PlatformInfo{K8SVersion: v.version}
		assert.Equal(t, v.major, info.K8SMajorVersion(), "K8SMajorVersion mismatch")
		assert.Equal(t, v.minor, info.K8SMinorVersion(), "K8SMinorVersion mismatch")
	}
}

func TestPlatformInfo_String(t *testing.T) {

	info := PlatformInfo{Name: OpenShift, K8SVersion: "456", OS: "foo/bar"}

	assert.Equal(t, "PlatformInfo [Name: OpenShift, K8SVersion: 456, OS: foo/bar]",
		info.String(), "PlatformInfo String() yields malformed result of %s", info.String())
}

func TestVersionHelpers(t *testing.T) {

	ocpTestVersions := []struct {
		version string
		major   string
		minor   string
		build   string
	}{
		{"3.11.69", "3", "11", "69"},
		{"4.1.0-rc.1", "4", "1", "0-rc.1"},
		{"1.2.3.4.5.6", "1", "2", "3.4.5.6"},
		{"1.2", "1", "2", ""},
	}

	for _, v := range ocpTestVersions {

		info := OpenShiftVersion{Version: v.version}
		assert.Equal(t, v.major, info.MajorVersion(), "OCPMajorVersion mismatch")
		assert.Equal(t, v.minor, info.MinorVersion(), "OCPMinorVersion mismatch")
		assert.Equal(t, v.build, info.BuildVersion(), "OCPBuildVersion mismatch")
	}
}

func TestOpenShiftVersion_String(t *testing.T) {

	info := OpenShiftVersion{Version: "1.1.1+"}

	assert.Equal(t, "OpenShiftVersion [Version: 1.1.1+]",
		info.String(), "OpenShiftVersion String() yields malformed result of %s", info.String())
}
