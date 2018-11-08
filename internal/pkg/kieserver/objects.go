package kieserver

import (
	"github.com/imdario/mergo"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/shared"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	"github.com/sirupsen/logrus"
)

func ConstructObject(object opv1.CustomObject, cr *opv1.App) opv1.CustomObject {
	for dcIndex, dc := range object.DeploymentConfigs {
		for containerIndex, c := range dc.Spec.Template.Spec.Containers {
			c.Env = shared.EnvOverride(cr.Spec.Objects.Server.Env, c.Env)

			err := mergo.Merge(&c.Resources, cr.Spec.Objects.Server.Resources, mergo.WithOverride)
			if err != nil {
				logrus.Error(err)
			}
			dc.Spec.Template.Spec.Containers[containerIndex] = c
		}
		object.DeploymentConfigs[dcIndex] = dc
	}
	return object
}
