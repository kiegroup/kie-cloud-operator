package kieserver

import (
	"github.com/imdario/mergo"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("kieapp.kieserver")

func ConstructObject(object v1.CustomObject, cr *v1.KieApp) v1.CustomObject {
	for dcIndex, dc := range object.DeploymentConfigs {
		for containerIndex, c := range dc.Spec.Template.Spec.Containers {
			c.Env = shared.EnvOverride(c.Env, cr.Spec.Objects.Server.Env)

			err := mergo.Merge(&c.Resources, cr.Spec.Objects.Server.Resources, mergo.WithOverride)
			if err != nil {
				log.Error(err, "Error merging interfaces")
			}
			dc.Spec.Template.Spec.Containers[containerIndex] = c
		}
		object.DeploymentConfigs[dcIndex] = dc
	}
	return object
}
