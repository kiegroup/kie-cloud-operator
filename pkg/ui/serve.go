package ui

//go:generate go run ../controller/kieapp/defaults/.packr/packr.go

import (
	"encoding/json"
	"io/ioutil"

	"github.com/RHsyseng/console-cr-form/pkg/web"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/gobuffalo/packr/v2"
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
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
	kieApp := &v1.KieApp{}
	err := yaml.Unmarshal([]byte(cr), kieApp)
	if err != nil {
		log.Errorf("Failed to parse CR based on %s, error is %v", cr, err)
		return err
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error("Failed to get in-cluster config", err)
		return err
	}
	err = v1.SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error("Failed to add scheme", err)
		return err
	}
	config.ContentConfig.GroupVersion = &v1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	restClient, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		log.Error("Failed to get REST client", err)
		return err
	}
	kieApp.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("KieApp"))
	err = restClient.Post().Namespace(getCurrentNamespace()).Body(kieApp).Resource("kieapps").Do().Into(kieApp)
	if err != nil {
		log.Error("Failed to create KIE app", err)
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
	Validation CustomResourceDefinitionValidation `json:"validation,omitempty"`
}

type CustomResourceDefinitionValidation struct {
	OpenAPIV3Schema spec.Schema `json:"openAPIV3Schema,omitempty"`
}

func getSchema() spec.Schema {
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
	return crd.Spec.Validation.OpenAPIV3Schema
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
	return v1.SchemeGroupVersion.String()
}
