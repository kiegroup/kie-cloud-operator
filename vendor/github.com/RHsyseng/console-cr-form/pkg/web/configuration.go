package web

import (
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	aggregate "k8s.io/apimachinery/pkg/util/errors"
)

type Configuration interface {
	Host() string
	Port() int
	Schema() spec.Schema
	ApiVersion() string
	Kind() string
	Form() Form
	CallBack(yaml string) error
}

func NewConfiguration(host string, port int, schema spec.Schema, apiVersion string, kind string, form Form, callback func(yaml string) error) (Configuration, error) {
	var errs []error
	if port == 0 {
		port = 8080
	}
	if apiVersion == "" {
		errs = append(errs, errors.New("No apiVersion value provided"))
	}
	if kind == "" {
		errs = append(errs, errors.New("No kind value provided"))
	}
	if len(form.Pages) == 0 {
		errs = append(errs, errors.New("No valid form provided"))
	}
	if callback == nil {
		errs = append(errs, errors.New("No callback provided"))
	}
	if len(errs) == 0 {
		return &configuration{host, port, schema, apiVersion, kind, form, callback}, nil
	}
	logrus.Debug("Configuration is invalid", errs)
	return nil, aggregate.NewAggregate(errs)
}

type configuration struct {
	host       string
	port       int
	schema     spec.Schema
	apiVersion string
	kind       string
	form       Form
	callback   func(yamlString string) error
}

func (config *configuration) Host() string {
	return config.host
}

func (config *configuration) Port() int {
	return config.port
}

func (config *configuration) Schema() spec.Schema {
	return config.schema
}

func (config *configuration) ApiVersion() string {
	return config.apiVersion
}

func (config *configuration) Kind() string {
	return config.kind
}

func (config *configuration) Form() Form {
	return config.form
}

func (config *configuration) CallBack(yaml string) error {
	return config.callback(yaml)
}
