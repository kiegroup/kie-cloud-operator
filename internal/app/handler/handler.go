package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/internal/pkg/defaults"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/kieserver"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.App:
		var objects []runtime.Object

		if event.Deleted {
			logrus.Infof("Deleting %s %s", o.Kind, o.Name)
			return nil
		}

		// further work required to support CR object updates
		checkUpdateStatus(o)

		if o.Status != "Installed" {
			env := o.Spec.Environment
			switch env {
			case "trial-ephemeral":
				logrus.Infof("Will set up a trial environment")
				objects = NewTrialEnv(o)
			case "authoring":
				logrus.Infof("Will set up an authoring environment")
				objects = NewAuthoringEnv(o)
			default:
				logrus.Infof("Environment is %s and not sure what to do with that!", env)
				return nil
			}

			for _, obj := range objects {
				var err error
				if o.Status == "Updating" {
					// when this is functional, resourceVersion for each object will need to be known/set - maybe attach to CR status as created?
					err = sdk.Update(obj)
				} else {
					err = sdk.Create(obj)
				}
				if err != nil {
					if errors.IsAlreadyExists(err) {
						logrus.Debugf("%s already exists, will not be created", obj.GetObjectKind().GroupVersionKind().Kind)
					} else {
						logrus.Errorf("Failed to create object : %v", err)
						bytes, err1 := json.Marshal(obj)
						if err1 != nil {
							logrus.Infof("Can't serialize", obj)
						} else {
							logrus.Infof("Object is ", string(bytes))
						}
						return err
					}
				}
			}

			// Update CR
			o.Status = "Installed"
			err := sdk.Update(o)
			if err != nil {
				return fmt.Errorf("failed to update %s status: %v", o.Kind, err)
			}
			logrus.Infof("%s %s is now Installed", o.Kind, o.Name)
		}
	}
	return nil
}

func NewTrialEnv(cr *v1.App) []runtime.Object {
	env := defaults.GetTrialEnvironment()
	console := rhpamcentr.ConstructObjects(env.Console, cr)
	server := kieserver.ConstructObjects(env.Servers[0], cr)
	return []runtime.Object{&console.DeploymentConfig, &console.Service, &console.Route, &server.DeploymentConfig, &server.Service, &server.Route}
}

func NewAuthoringEnv(cr *v1.App) []runtime.Object {
	return []runtime.Object{}
}

// figure out later how to know if there is an update to CR, and mark it's status as Updated
func checkUpdateStatus(o *v1.App) {
	/*
		if o.Status != "" {
			if !reflect.DeepEqual(??.Spec, o.Spec) {
			}
			logrus.Infof("Updating %s %s", o.Kind, o.Name)
			o.Status = "Updating"
		}
	*/
}
