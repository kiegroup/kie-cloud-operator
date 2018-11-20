package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(mgr manager.Manager) error {
	for _, functions := range AddToManagerFuncs {
		if err := functions(mgr); err != nil {
			return err
		}
	}
	return nil
}
