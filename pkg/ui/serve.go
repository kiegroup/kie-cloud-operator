package ui

//go:generate go run -mod=vendor ../controller/kieapp/defaults/.packr/packr.go

import (
	"encoding/json"
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"io/ioutil"

	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/gobuffalo/packr/v2"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var log = logs.GetLogger("ui")

// Listen ...
func Listen() {
	config, err := web.NewConfiguration("", 8080, getSchema(), getApiVersion(), getObjectKind(), getForm(), apply)
	if err != nil {
		log.Fatal("Failed to configure web server", err)
	}
	if err := web.RunWebServer(config); err != nil {
		log.Fatal("Failed to run web server", err)
	}
}

func apply(cr string) error {
	log.Debugf("Will deploy KieApp based on yaml %v", cr)
	kieApp := &api.KieApp{}
	err := yaml.Unmarshal([]byte(cr), kieApp)
	if err != nil {
		log.Debugf("Failed to parse CR based on %s. Cause: ", cr, err)
		return err
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Debug("Failed to get in-cluster config", err)
		return err
	}
	err = api.SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Debug("Failed to add scheme", err)
		return err
	}
	config.ContentConfig.GroupVersion = &api.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	restClient, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		log.Debug("Failed to get REST client", err)
		return err
	}
	kieApp.SetGroupVersionKind(api.SchemeGroupVersion.WithKind("KieApp"))
	err = restClient.Post().Namespace(getCurrentNamespace()).Body(kieApp).Resource("kieapps").Do().Into(kieApp)
	if err != nil {
		log.Debug("Failed to create KIE app", err)
		return err
	}
	log.Infof("Created KIE application named %s", kieApp.Name)
	return nil
}

func getCurrentNamespace() string {
	bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatal("Failed to read current namespace and cannot proceed ", err)
	}
	return string(bytes)
}

type CustomResourceDefinition struct {
	Spec CustomResourceDefinitionSpec `json:"spec,omitempty"`
}

type CustomResourceDefinitionSpec struct {
	Versions   []CustomResourceDefinitionVersion  `json:"versions,omitempty"`
	Validation CustomResourceDefinitionValidation `json:"validation,omitempty"`
}

type CustomResourceDefinitionVersion struct {
	Name   string                             `json:"Name,omitempty"`
	Schema CustomResourceDefinitionValidation `json:"schema,omitempty"`
}

type CustomResourceDefinitionValidation struct {
	OpenAPIV3Schema spec.Schema `json:"openAPIV3Schema,omitempty"`
}

func getSchema() spec.Schema {
	schema := spec.Schema{}
	box := packr.New("CRD", "../../deploy/crds/")
	yamlByte, err := box.Find("kieapp.crd.yaml")
	if err != nil {
		log.Fatal(err)
		panic("Failed to retrieve crd, there must be an environment problem!")
	}
	crd := &CustomResourceDefinition{}
	err = yaml.Unmarshal(yamlByte, crd)
	if err != nil {
		log.Fatal(err)
		panic("Failed to unmarshal static schema, there must be an environment problem!")
	}
	for _, v := range crd.Spec.Versions {
		if v.Name == api.SchemeGroupVersion.Version {
			schema = v.Schema.OpenAPIV3Schema
		}
	}
	return schema
}

func getForm() web.Form {
	box := packr.New("form", "../../deploy/ui/")
	jsonBytes, err := box.Find("form.json")
	if err != nil {
		log.Fatal(err)
		panic("Failed to retrieve ui form, there must be an environment problem!")
	}
	form := &web.Form{}
	err = json.Unmarshal(jsonBytes, form)
	if err != nil {
		log.Fatal(err)
		panic("Failed to unmarshal static ui form, there must be an environment problem!")
	}
	return *form
}

func getObjectKind() string {
	box := packr.New("CRD", "../../deploy/crds/")
	yamlByte, err := box.Find("kieapp.crd.yaml")
	if err != nil {
		log.Fatal(err)
		panic("Failed to retrieve crd, there must be an environment problem!")
	}
	crd := &v1beta1.CustomResourceDefinition{}
	err = yaml.Unmarshal(yamlByte, crd)
	if err != nil {
		panic("Failed to unmarshal static schema, there must be an environment problem!")
	}
	return crd.Spec.Names.Kind
}

func getApiVersion() string {
	return api.SchemeGroupVersion.String()
}
