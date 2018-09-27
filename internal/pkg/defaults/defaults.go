package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
)

func GetTrialEnvironment() v1.Environment {
	env := v1.Environment{}
	loadYaml("trial-env.yaml", &env)
	return env
}

func GetConsoleObject() v1.OpenShiftObject {
	object := v1.OpenShiftObject{}
	loadYaml("console.yaml", &object)
	return object
}

func GetServerObject() v1.OpenShiftObject {
	object := v1.OpenShiftObject{}
	loadYaml("server.yaml", &object)
	return object
}

func loadYaml(filename string, o interface{}) {
	box := packr.NewBox("../../../config/app")
	yaml.Unmarshal(box.Bytes(filename), &o)
}
