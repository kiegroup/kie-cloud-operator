package rhpamcentr

import (
	"github.com/imdario/mergo"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/shared"
)

var log = logs.GetLogger("kieapp.rhpamcentr")

func ConstructObject(object v1.CustomObject, cr *v1.KieApp) v1.CustomObject {
	for dcIndex, dc := range object.DeploymentConfigs {
		for containerIndex, c := range dc.Spec.Template.Spec.Containers {
			c.Env = shared.EnvOverride(c.Env, cr.Spec.Objects.Console.Env)

			err := mergo.Merge(&c.Resources, cr.Spec.Objects.Console.Resources, mergo.WithOverride)
			if err != nil {
				log.Error(err, "Error merging interfaces")
			}
			dc.Spec.Template.Spec.Containers[containerIndex] = c
		}
		object.DeploymentConfigs[dcIndex] = dc
	}
	return object
}
