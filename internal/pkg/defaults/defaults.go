package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/rhpam/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
)

func GetTrialEnvironment() v1alpha1.Environment {
	env := v1alpha1.Environment{}
	loadYaml("trial-env.yaml", &env)
	return env
}

func GetConsoleObject() v1alpha1.OpenShiftObject {
	object := v1alpha1.OpenShiftObject{}
	loadYaml("console.yaml", &object)
	return object
}

func GetServerObject() v1alpha1.OpenShiftObject {
	object := v1alpha1.OpenShiftObject{}
	loadYaml("server.yaml", &object)
	return object
}

func loadYaml(filename string, o interface{}) {
	box := packr.NewBox("../../../config/app")
	yaml.Unmarshal(box.Bytes(filename), &o)
}
