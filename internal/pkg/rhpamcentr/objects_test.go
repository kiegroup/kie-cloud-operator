package rhpamcentr

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
	bc := GetRHMAPCentr(cr)
	assert.Equal(t, len(bc), 3, "Expect the workbench to have 3 objects")

	assert.IsType(t, &v1.DeploymentConfig{}, bc[0], "Expected the first object to be a deployment config")
	dc := bc[0].(*v1.DeploymentConfig)
	logrus.Debugf("dc is %v", dc)
	assert.Equal(t, dc.Spec.Template.Spec.Containers[0].Name, appname+"-"+constants.RhpamcentrServicePrefix, "Deployment config container name not as expected")

	assert.IsType(t, &corev1.Service{}, bc[1], "Expected the second object to be a service")
	service := bc[1].(*corev1.Service)
	logrus.Debugf("service is %v", service)
	assert.Equal(t, len(service.Spec.Ports), 2, "Service should expose 2 ports")
	ports := []int{int(service.Spec.Ports[0].Port), int(service.Spec.Ports[1].Port)}
	sort.Ints(ports)
	assert.Equal(t, ports[0], 8001, "Expected the lower port to be 8001")
	assert.Equal(t, ports[1], 8080, "Expected the higher port to be 8080")

	assert.IsType(t, &routev1.Route{}, bc[2], "Expected the third object to be a route")
	openshiftRoute := bc[2].(*routev1.Route)
	logrus.Debugf("route is %v", openshiftRoute)
	assert.Equal(t, openshiftRoute.Spec.To.Name, appname+"-"+constants.RhpamcentrServicePrefix, "Route name not as expected")
}
