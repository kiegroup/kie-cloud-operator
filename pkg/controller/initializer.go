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
			version := openshift.MapKnownVersion(info)
			if version.Version != "" {
				log.Info(fmt.Sprintf("OpenShift Version: %s", version.Version))
				reconciler.OcpVersion = version.Version
				reconciler.OcpVersionMajor = version.MajorVersion()
				reconciler.OcpVersionMinor = version.MinorVersion()
			}
		}

		// ??? check against supported ocp versions here and log something if version doesn't match???

		kieapp.CreateConsoleYAMLSamples(&reconciler)
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
