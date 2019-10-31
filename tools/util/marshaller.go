package util

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func MarshallObject(obj interface{}, writer io.Writer) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var r unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &r.Object); err != nil {
		return err
	}

	// remove status and metadata.creationTimestamp
	unstructured.RemoveNestedField(r.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "spec", "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "status")

	// remove dataSource from PVCs if empty
	templates, exists, err := unstructured.NestedSlice(r.Object, "spec", "dataVolumeTemplates")
	if exists {
		for _, tmpl := range templates {
			template := tmpl.(map[string]interface{})
			_, exists, err = unstructured.NestedString(template, "spec", "pvc", "dataSource")
			if !exists {
				unstructured.RemoveNestedField(template, "spec", "pvc", "dataSource")
			}
		}
		unstructured.SetNestedSlice(r.Object, templates, "spec", "dataVolumeTemplates")
	}
	objects, exists, err := unstructured.NestedSlice(r.Object, "objects")
	if exists {
		for _, obj := range objects {
			object := obj.(map[string]interface{})
			kind, exists, _ := unstructured.NestedString(object, "kind")
			if exists && kind == "PersistentVolumeClaim" {
				_, exists, err = unstructured.NestedString(object, "spec", "dataSource")
				if !exists {
					unstructured.RemoveNestedField(object, "spec", "dataSource")
				}
			}
		}
		unstructured.SetNestedSlice(r.Object, objects, "objects")
	}

	deployments, exists, err := unstructured.NestedSlice(r.Object, "spec", "install", "spec", "deployments")
	if exists {
		for _, obj := range deployments {
			deployment := obj.(map[string]interface{})
			unstructured.RemoveNestedField(deployment, "metadata", "creationTimestamp")
			unstructured.RemoveNestedField(deployment, "spec", "template", "metadata", "creationTimestamp")
			unstructured.RemoveNestedField(deployment, "status")
		}
		unstructured.SetNestedSlice(r.Object, deployments, "spec", "install", "spec", "deployments")
	}

	// remove "managed by operator" label...
	labels, exists, err := unstructured.NestedMap(r.Object, "metadata", "labels")
	if exists {
		//delete(labels, v1.ManagedByLabel)
		unstructured.SetNestedMap(r.Object, labels, "metadata", "labels")
	}

	jsonBytes, err = json.Marshal(r.Object)
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return err
	}

	// fix templates by removing unneeded single quotes...
	s := string(yamlBytes)
	s = strings.Replace(s, "'{{", "{{", -1)
	s = strings.Replace(s, "}}'", "}}", -1)

	// fix double quoted strings by removing unneeded single quotes...
	s = strings.Replace(s, " '\"", " \"", -1)
	s = strings.Replace(s, "\"'\n", "\"\n", -1)

	yamlBytes = []byte(s)

	_, err = writer.Write(yamlBytes)
	if err != nil {
		return err
	}

	return nil
}

func RawMessagePointer(str string) *json.RawMessage {
	message := json.RawMessage{}
	message = []byte(str)
	return &message
}
