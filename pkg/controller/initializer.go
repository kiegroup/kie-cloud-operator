package controller

import (
	"fmt"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
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
				log.Info(fmt.Sprintf("OpenShift Version: %s", mappedVersion.Version))
				reconciler.OcpVersion = mappedVersion.Version
				reconciler.OcpVersionMajor = mappedVersion.MajorVersion()
				reconciler.OcpVersionMinor = mappedVersion.MinorVersion()
				/* ?? warning if ocp version isn't in SupportedOcpVersions slice ??
				if _, ok := shared.Find(constants.SupportedOcpVersions, reconciler.OcpVersion); !ok {
					log.Warn("OpenShift version not supported.")
				}
				*/
			} else {
				log.Warn("OpenShift version could not be determined.")
			}
		}
		kieapp.CreateConsoleYAMLSamples(&reconciler)
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
