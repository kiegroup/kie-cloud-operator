package kieapp

//go:generate go run -mod=vendor ./defaults/.packr/packr.go

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
)

func CreateConsoleYAMLSamples(r *Reconciler) {
	log.Info("Loading CR YAML samples.")
	box := packr.New("cryamlsamples", "../../../deploy/crs")
	if box.List() == nil {
		log.Warnf("CR YAML folder is empty. It is not loaded.")
		return
	}

	resMap := make(map[string]string)
	for _, filename := range box.List() {
		yamlStr, err := box.FindString(filename)
		if err != nil {
			resMap[filename] = err.Error()
			continue
		}
		kieApp := api.KieApp{}
		err = yaml.Unmarshal([]byte(yamlStr), &kieApp)
		if err != nil {
			resMap[filename] = err.Error()
			continue
		}
		yamlSample, err := openshift.GetConsoleYAMLSample(&kieApp)
		if err != nil {
			resMap[filename] = err.Error()
			continue
		}
		err = r.Service.Create(context.TODO(), yamlSample)
		if err != nil {
			resMap[filename] = err.Error()
			continue
		}
		resMap[filename] = "Applied"
	}

	for k, v := range resMap {
		log.Infof("yaml file: %s %s ", k, v)
	}
}
