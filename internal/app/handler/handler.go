package handler

import (
	"context"
	"encoding/json"

	"github.com/kiegroup/kie-cloud-operator/internal/pkg/defaults"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/kieserver"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"

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
	case *opv1.App:
		var err error

		if event.Deleted {
			logrus.Infof("Deleting %s %s", o.Name, o.Kind)
			return nil
		}

		// further work required to support CR object updates
		checkUpdateStatus(o)

		if (o.Status != "Installed") && (o.Status != "Error") {
			var objects []runtime.Object
			if o.Spec.Environment != "" {
				logrus.Infof("Will set up %s environment", o.Spec.Environment)

				objects, err = NewEnv(o)
				if err != nil {
					o.Status = "Error"
					logrus.Error(err)
				}
			}

			for _, obj := range objects {
				var err error
				if o.Status == "Updating" {
					// when this is functional, resourceVersion for each object will need to be known/set - maybe attach to CR status as created?
					err = sdk.Update(obj)
				} else if o.Status != "Error" {
					o.Status = "Installed"
					err = sdk.Create(obj)
				}
				if err != nil {
					if errors.IsAlreadyExists(err) {
						logrus.Warnf("%s already exists, will not be created", obj.GetObjectKind().GroupVersionKind().Kind)
					} else {
						logrus.Errorf("Failed to create object %v: %v", obj, err)
						bytes, err1 := json.Marshal(obj)
						if err1 != nil {
							logrus.Infof("Can't serialize", obj)
						} else {
							logrus.Infof("Object is ", string(bytes))
						}
						return nil
					}
				}
			}

			// Update CR
			err := sdk.Update(o)
			if err != nil {
				logrus.Errorf("failed to update %s status: %v", o.Kind, err)
			}
			if o.Status != "Error" {
				logrus.Infof("%s %s is now installed", o.Name, o.Kind)
			}
		}
	}
	return nil
}

func NewEnv(cr *opv1.App) ([]runtime.Object, error) {
	var objs []runtime.Object
	env, err := defaults.GetEnvironment(cr)
	if err != nil {
		return []runtime.Object{}, err
	}

	// create object slice for deployment
	objs = shared.ObjectAppend(objs, rhpamcentr.ConstructObject(env.Console, cr))
	for _, s := range env.Servers {
		objs = shared.ObjectAppend(objs, kieserver.ConstructObject(s, cr))
	}

	objs = shared.SetReferences(objs, cr)

	return objs, err
}

// figure out later how to know if there is an update to CR, and mark it's status as Updated
func checkUpdateStatus(o *opv1.App) {
	/*
		if o.Status != "" {
			if !reflect.DeepEqual(??.Spec, o.Spec) {
			}
			logrus.Infof("Updating %s %s", o.Kind, o.Name)
			o.Status = "Updating"
		}
	*/
}
