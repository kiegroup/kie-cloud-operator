package controller

import (
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logs.GetLogger("kie-cloud-operator.controller")

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	addManager := func(mgr manager.Manager) error {
		k8sService := kubernetes.GetInstance(mgr)
		reconciler := kieapp.Reconciler{Service: &k8sService}
		return kieapp.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager.Manager) error{addManager}
}
