package controller

import (
	"fmt"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	"golang.org/x/mod/semver"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logs.GetLogger("kieapp.initializer")

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	addManager := func(mgr manager.Manager) error {
		k8sService := kubernetes.GetInstance(mgr)
		reconciler := kieapp.Reconciler{Service: &k8sService}
		info, err := openshift.GetPlatformInfo(mgr.GetConfig())
		if err != nil {
			log.Error(err)
		}
		if info.IsOpenShift() {
			mappedVersion := openshift.MapKnownVersion(info)
			if mappedVersion.Version != "" {
				if _, ok := shared.Find(constants.SupportedOcpVersions, mappedVersion.Version); !ok {
					log.Warn("OpenShift version not supported.")
				}
				reconciler.OcpVersion = semver.MajorMinor("v" + mappedVersion.Version)
				log.Info(fmt.Sprintf("OpenShift Version: %s", reconciler.OcpVersion))
			} else {
				log.Warn("OpenShift version could not be determined.")
			}
		}
		if semver.Compare(reconciler.OcpVersion, "v4.3") >= 0 || reconciler.OcpVersion == "" {
			kieapp.CreateConsoleYAMLSamples(&reconciler)
		}
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
