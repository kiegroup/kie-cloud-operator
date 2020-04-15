package controller

import (
	"fmt"

	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/blang/semver"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
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
				if _, ok := shared.Find(constants.SupportedOcpVersions, mappedVersion.Version); !ok {
					log.Warn("OpenShift version not supported.")
				}
				if v, err := semver.New(mappedVersion.Version + ".0"); err == nil {
					reconciler.OcpVersion = v
				} else {
					log.Warn("OpenShift version could not be parsed.")
				}
			} else {
				log.Warn("OpenShift version could not be determined.")
			}
			if reconciler.OcpVersion == nil || reconciler.OcpVersion.GE(semver.MustParse("4.3.0")) {
				kieapp.CreateConsoleYAMLSamples(&reconciler)
			}
		}
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
