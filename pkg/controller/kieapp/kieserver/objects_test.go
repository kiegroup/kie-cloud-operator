package kieserver

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/constants"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestConstructServerObject(t *testing.T) {
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	}
	envAddition := corev1.EnvVar{
		Name:  "SERVER_TEST",
		Value: "test",
	}
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.KieAppSpec{
			Environment:    "trial",
			KieDeployments: 3,
			Objects: v1.KieAppObjects{
				Server: v1.KieAppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}

	env, common, err := defaults.GetEnvironment(cr, fake.NewFakeClient())
	assert.Nil(t, err)

	var objects []v1.CustomObject
	for _, s := range env.Servers {
		object := ConstructObject(s, common, cr)
		objects = append(objects, object)
	}
	assert.Equal(t, cr.Spec.KieDeployments, len(env.Servers))
	assert.Equal(t, fmt.Sprintf("%s-kieserver-%d", cr.Name, cr.Spec.KieDeployments-1), env.Servers[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-kieserver-openshift:%s", strings.Join(re.FindAllString(constants.RhpamVersion, -1), ""), constants.ImageStreamTag), env.Servers[0].DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, objects[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, objects[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
	assert.Contains(t, objects[cr.Spec.KieDeployments-1].DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	})
}
