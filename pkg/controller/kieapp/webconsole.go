package kieapp

//go:generate go run -mod=vendor ./defaults/.packr/packr.go

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/RHsyseng/operator-utils/pkg/utils/openshift"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/version"
	consolev1 "github.com/openshift/api/console/v1"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/mod/semver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (reconciler *Reconciler) createConsoleYAMLSamples() error {
	if semver.Compare(reconciler.OcpVersion, "v4.3") >= 0 || reconciler.OcpVersion == "" {
		box := packr.New("cryamlsamples", "../../../deploy/crs/v2")
		if box.List() == nil {
			return errors.New("cr yaml folder is empty, not loaded")
		}
		for _, filename := range box.List() {
			yamlStr, err := box.FindString(filename)
			if err != nil {
				return err
			}
			kieApp := api.KieApp{}
			err = yaml.Unmarshal([]byte(yamlStr), &kieApp)
			if err != nil {
				return err
			}
			yamlSample, err := openshift.GetConsoleYAMLSample(&kieApp)
			if err != nil {
				return err
			}
			yamlSample.Name = strings.ToLower(kieApp.GetObjectKind().GroupVersionKind().Kind) + "-" + yamlSample.Name
			yamlSample.SetGroupVersionKind(consolev1.SchemeGroupVersion.WithKind("ConsoleYAMLSample"))
			log := log.With("kind", yamlSample.Kind, "name", yamlSample.Name)
			yamlSample.Spec.YAML = consolev1.ConsoleYAMLSampleYAML(
				fmt.Sprintf(`apiVersion: %s
kind: %s
metadata:
  name: %s
%s`,
					kieApp.GetObjectKind().GroupVersionKind().GroupVersion().String(),
					kieApp.GetObjectKind().GroupVersionKind().Kind,
					kieApp.Name,
					getSpecString(yamlStr),
				),
			)
			if yamlSample.Spec.Snippet {
				yamlSample.Spec.YAML = consolev1.ConsoleYAMLSampleYAML(fmt.Sprintf(`%s`, getSpecString(yamlStr)))
			}
			newSample := &consolev1.ConsoleYAMLSample{}
			err = reconciler.Service.Get(context.TODO(), types.NamespacedName{Name: yamlSample.Name}, newSample)
			if apierrors.IsNotFound(err) {
				yamlSample.SetAnnotations(map[string]string{api.SchemeGroupVersion.Group: version.Version})
				err = reconciler.Service.Create(context.TODO(), yamlSample)
				if err != nil {
					return err
				}
				log.Info("Created")
			} else if err == nil {
				if newSample.GetAnnotations() == nil ||
					semver.Compare(semver.MajorMinor("v"+version.Version), "v"+newSample.GetAnnotations()[api.SchemeGroupVersion.Group]) > 0 {
					newSample.SetAnnotations(map[string]string{api.SchemeGroupVersion.Group: version.Version})
					newSample.Spec = yamlSample.Spec
					err := reconciler.Service.Update(context.TODO(), newSample)
					if err != nil {
						return err
					}
					log.Info("Updated")
				}
			} else {
				return err
			}
		}
		return nil
	}
	return errors.New("console yaml samples not installed, incompatible ocp version")
}

func getSpecString(yamlStr string) string {
	jsonObject, err := yaml.YAMLToJSON([]byte(yamlStr))
	if err != nil {
		log.Error(err)
		return ""
	}
	jsonStr, err := sjson.Set(`{"spec":{}}`, "spec", gjson.GetBytes(jsonObject, "spec").Value())
	if err != nil {
		log.Error(err)
		return ""
	}
	yamlObject, err := yaml.JSONToYAML([]byte(jsonStr))
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(yamlObject)
}
