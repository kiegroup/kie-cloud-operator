package kieapp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/kiegroup/kie-cloud-operator/version"
	routev1 "github.com/openshift/api/route/v1"
	operators "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestUpdateLink(t *testing.T) {
	service := test.MockServiceWithExtraScheme(&operators.ClusterServiceVersion{}, &appsv1.Deployment{}, &corev1.Pod{})

	opMajor, opMinor, _ := defaults.MajorMinorMicro(version.Version)
	box := packr.New("CSV", "../../../deploy/catalog_resources/redhat/"+opMajor+"."+opMinor)
	bytes, err := box.Find("businessautomation-operator." + version.Version + ".clusterserviceversion.yaml")
	assert.Nil(t, err, "Error reading CSV file")
	csv := &operators.ClusterServiceVersion{}
	err = yaml.Unmarshal(bytes, csv)
	assert.Nil(t, err, "Error parsing CSV file")
	assert.False(t, strings.Contains(csv.Spec.Description, constants.ConsoleDescription), "Should be no information about link in description")

	err = service.Create(context.TODO(), csv)
	assert.Nil(t, err, "Error creating the CSV")

	box = packr.New("Operator", "../../../deploy")
	bytes, err = box.Find("operator.yaml")
	assert.Nil(t, err, "Error reading Operator file")
	operator := &appsv1.Deployment{}
	err = yaml.Unmarshal(bytes, operator)
	assert.Nil(t, err, "Error parsing Operator file")

	operator.Namespace = "placeholder"
	err = controllerutil.SetControllerReference(csv, operator, service.GetScheme())
	assert.Nil(t, err, "Error setting operator owner as CSV")

	err = service.Create(context.TODO(), operator)
	assert.Nil(t, err, "Error creating the Operator")

	var url string
	service.CreateFunc = func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
		if route, matched := obj.(*routev1.Route); matched {
			url = fmt.Sprintf("%s.apps.example.com", route.Name)
			route.Spec.Host = url
		}
		return service.Client.Create(ctx, obj, opts...)
	}
	deployConsole(&Reconciler{Service: service}, operator)

	updatedCSV := &operators.ClusterServiceVersion{}
	err = service.Get(context.TODO(), types.NamespacedName{Name: csv.Name, Namespace: csv.Namespace}, updatedCSV)
	assert.Nil(t, err, "Error fetching CSV from client")

	link := getConsoleLink(updatedCSV)
	assert.NotNil(t, link, "Found no console link in CSV")
	assert.True(t, strings.Contains(updatedCSV.Spec.Description, constants.ConsoleDescription), "Found no information about link in description")
	assert.Equal(t, fmt.Sprintf("https://%s", url), link.URL, "The console link did not have the expected value")
}

func TestUpdateExistingLink(t *testing.T) {
	service := test.MockServiceWithExtraScheme(&operators.ClusterServiceVersion{}, &appsv1.Deployment{}, &corev1.Pod{})

	opMajor, opMinor, _ := defaults.MajorMinorMicro(version.Version)
	box := packr.New("CSV", "../../../deploy/catalog_resources/redhat/"+opMajor+"."+opMinor)
	bytes, err := box.Find("businessautomation-operator." + version.Version + ".clusterserviceversion.yaml")
	assert.Nil(t, err, "Error reading CSV file")
	csv := &operators.ClusterServiceVersion{}
	err = yaml.Unmarshal(bytes, csv)
	assert.Nil(t, err, "Error parsing CSV file")
	assert.False(t, strings.Contains(csv.Spec.Description, constants.ConsoleDescription), "Should be no information about link in description")

	csv.Spec.Links = append([]operators.AppLink{{Name: constants.ConsoleLinkName, URL: "some-bad-link"}}, csv.Spec.Links...)

	err = service.Create(context.TODO(), csv)
	assert.Nil(t, err, "Error creating the CSV")

	box = packr.New("Operator", "../../../deploy")
	bytes, err = box.Find("operator.yaml")
	assert.Nil(t, err, "Error reading Operator file")
	operator := &appsv1.Deployment{}
	err = yaml.Unmarshal(bytes, operator)
	assert.Nil(t, err, "Error parsing Operator file")

	operator.Namespace = "placeholder"
	err = controllerutil.SetControllerReference(csv, operator, service.GetScheme())
	assert.Nil(t, err, "Error setting operator owner as CSV")

	err = service.Create(context.TODO(), operator)
	assert.Nil(t, err, "Error creating the Operator")

	var url string
	service.CreateFunc = func(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
		if route, matched := obj.(*routev1.Route); matched {
			url = fmt.Sprintf("%s.apps.example.com", route.Name)
			route.Spec.Host = url
		}
		return service.Client.Create(ctx, obj, opts...)
	}
	deployConsole(&Reconciler{Service: service}, operator)

	updatedCSV := &operators.ClusterServiceVersion{}
	err = service.Get(context.TODO(), types.NamespacedName{Name: csv.Name, Namespace: csv.Namespace}, updatedCSV)
	assert.Nil(t, err, "Error fetching CSV from client")

	link := getConsoleLink(updatedCSV)
	assert.NotNil(t, link, "Found no console link in CSV")
	assert.True(t, strings.Contains(updatedCSV.Spec.Description, constants.ConsoleDescription), "Found no information about link in description")
	assert.Equal(t, fmt.Sprintf("https://%s", url), link.URL, "The console link did not have the expected value")
}
