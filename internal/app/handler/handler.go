package handler

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/kiegroup/kie-cloud-operator/internal/pkg/kieserver"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/rhpamcentr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/rhpam/v1alpha1"
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
	case *v1alpha1.App:
		var objects []runtime.Object
		env := o.Spec.Environment
		switch env {
		case "trial-ephemeral":
			logrus.Infof("Will set up a trial environment")
			objects = newTrialEnv(o)
		case "authoring":
			logrus.Infof("Will set up an authoring environment")
			objects = newAuthoringEnv(o)
		default:
			logrus.Infof("Environment is %v and not sure what to do with that!", env)
			return nil
		}
		for _, pod := range objects {
			err := sdk.Create(pod)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					logrus.Debugf("%s already exists, will not be created", pod.GetObjectKind().GroupVersionKind().Kind)
				} else {
					logrus.Errorf("Failed to create pod : %v", err)
					bytes, err1 := json.Marshal(pod)
					if err1 != nil {
						fmt.Println("Can't serialize", pod)
					} else {
						fmt.Println("Pod is ", string(bytes))
					}
					return err
				}
			}
		}
	}
	return nil
}

func newTrialEnv(cr *v1alpha1.App) []runtime.Object {
	return append(rhpamcentr.GetRHMAPCentr(cr), kieserver.GetKieServer(cr)...)
}

func newAuthoringEnv(cr *v1alpha1.App) []runtime.Object {
	return []runtime.Object{}
}
