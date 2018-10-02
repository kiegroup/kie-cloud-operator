package defaults

//go:generate sh -c "CGO_ENABLED=0 go run .packr/packr.go $PWD"

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr"
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	"github.com/sirupsen/logrus"
)

func GetEnvironment(e string) (v1.Environment, error) {
	env := v1.Environment{}
	err := loadYaml(fmt.Sprintf("envs/%s.yaml", e), &env)
	if err != nil {
		return env, err
	}
	return env, nil
}

func GetConsoleObject() v1.CustomObject {
	object := v1.CustomObject{}
	err := loadYaml("console.yaml", &object)
	if err != nil {
		logrus.Errorln(err)
	}
	return object
}

func GetServerObject() v1.CustomObject {
	object := v1.CustomObject{}
	err := loadYaml("server.yaml", &object)
	if err != nil {
		logrus.Errorln(err)
	}
	return object
}

func loadYaml(filename string, o interface{}) error {
	box := packr.NewBox("../../../config/app")
	if box.Has(filename) {
		yaml.Unmarshal(box.Bytes(filename), &o)
	} else {
		return fmt.Errorf("%s does not exist, environment not deployed", filename)
	}
	return nil
}
