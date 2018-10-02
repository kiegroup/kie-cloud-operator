package kieserver

import (
	"sort"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/internal/constants"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	"github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTrialEnvironmentStructure(t *testing.T) {
	appname := "unittest"
	event := sdk.Event{
		Object: &opv1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name: appname,
			},
			Spec: opv1.AppSpec{
				Environment: "trial",
			},
		},
		Deleted: false}
	logrus.Debugf("Testing with environment %v", event.Object.(*opv1.App).Spec.Environment)
	cr := event.Object.(*opv1.App)
	bc := GetKieServer(cr)
	assert.Equal(t, len(bc), 3, "Expect the workbench to have 3 objects")

	assert.IsType(t, &v1.DeploymentConfig{}, bc[0], "Expected the first object to be a deployment config")
	dc := bc[0].(*v1.DeploymentConfig)
	logrus.Debugf("dc is %v", dc)
	assert.Equal(t, dc.Spec.Triggers[0].ImageChangeParams.From.Name, "rhpam70-kieserver-openshift:1.2")
	container := dc.Spec.Template.Spec.Containers[0]
	assert.Equal(t, container.Name, appname+"-"+constants.KieServerServicePrefix, "Deployment config container name not as expected")
	ports := []int{int(container.Ports[0].ContainerPort), int(container.Ports[1].ContainerPort)}
	sort.Ints(ports)
	assert.Equal(t, ports[0], 8080, "Expected the lower port to be 8080")
	assert.Equal(t, ports[1], 8778, "Expected the higher port to be 8778")
	assert.Equal(t, container.ReadinessProbe.InitialDelaySeconds, int32(60))
	assert.Equal(t, container.ReadinessProbe.TimeoutSeconds, int32(2))
	assert.Equal(t, container.ReadinessProbe.PeriodSeconds, int32(30))
	assert.Equal(t, container.ReadinessProbe.FailureThreshold, int32(6))
	assert.Equal(t, container.LivenessProbe.InitialDelaySeconds, int32(180))
	assert.Equal(t, container.LivenessProbe.TimeoutSeconds, int32(2))
	assert.Equal(t, container.LivenessProbe.PeriodSeconds, int32(15))
	assert.Equal(t, *container.Resources.Limits.Memory(), resource.MustParse("220Mi"))

	assert.IsType(t, &corev1.Service{}, bc[1], "Expected the second object to be a service")
	service := bc[1].(*corev1.Service)
	logrus.Debugf("service is %v", service)
	assert.Equal(t, len(service.Spec.Ports), 1, "Service should expose 1 port")
	assert.Equal(t, service.Spec.Ports[0].Port, int32(8080), "Expected the port to be 8080")

	assert.IsType(t, &routev1.Route{}, bc[2], "Expected the third object to be a route")
	openshiftRoute := bc[2].(*routev1.Route)
	logrus.Debugf("route is %v", openshiftRoute)
	assert.Equal(t, openshiftRoute.Spec.To.Name, appname+"-"+constants.KieServerServicePrefix, "Route name not as expected")
}
