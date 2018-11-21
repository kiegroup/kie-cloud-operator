package controller

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kieapp.Add)
}
