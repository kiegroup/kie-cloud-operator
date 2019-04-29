package web

import (
	"github.com/go-openapi/spec"
)

type Configuration interface {
	Host() string
	Port() int
	Schema() spec.Schema
	Form() Form
}

type ConfigurationHolder struct {
	HostField   string
	PortField   int
	SchemaField spec.Schema
	FormField   Form
}

func (config *ConfigurationHolder) Host() string {
	return config.HostField
}

func (config *ConfigurationHolder) Port() int {
	return config.PortField
}

func (config *ConfigurationHolder) Schema() spec.Schema {
	return config.SchemaField
}

func (config *ConfigurationHolder) Form() Form {
	return config.FormField
}
