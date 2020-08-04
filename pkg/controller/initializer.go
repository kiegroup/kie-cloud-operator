package controller

import (
	"fmt"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
	"golang.org/x/mod/semver"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logs.GetLogger("kieapp.initializer")

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	addManager := func(mgr manager.Manager) error {
		k8sService := kubernetes.GetInstance(mgr)
		reconciler := kieapp.Reconciler{Service: &k8sService, Status: mgr.GetClient().Status()}
		info, err := openshift.GetPlatformInfo(mgr.GetConfig())
		if err != nil {
			log.Error(err)
		}
		if info.IsOpenShift() {
			mappedVersion := openshift.MapKnownVersion(info)
			if mappedVersion.Version != "" {
				reconciler.OcpVersion = semver.MajorMinor("v" + mappedVersion.Version)
				log.Info(fmt.Sprintf("OpenShift Version: %s", reconciler.OcpVersion))
			} else {
				log.Warn("OpenShift version could not be determined.")
			}
		}
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
