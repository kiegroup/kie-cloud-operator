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
		env := o.Spec.Environment
		switch env {
		case "trial-ephemeral":
			logrus.Infof("Will set up a trial environment")
			objects = NewTrialEnv(o)
		case "authoring":
			logrus.Infof("Will set up an authoring environment")
			objects = NewAuthoringEnv(o)
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

func NewTrialEnv(cr *v1.App) []runtime.Object {
	env := defaults.GetTrialEnvironment()
	console := rhpamcentr.ConstructObjects(env.Console, cr)
	server := kieserver.ConstructObjects(env.Servers[0], cr)
	return []runtime.Object{&console.DeploymentConfig, &console.Service, &console.Route,&server.DeploymentConfig, &server.Service, &server.Route}
}

func NewAuthoringEnv(cr *v1.App) []runtime.Object {
	return []runtime.Object{}
}
