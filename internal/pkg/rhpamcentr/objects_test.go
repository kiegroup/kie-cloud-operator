package rhpamcentr

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kiegroup/kie-cloud-operator/internal/constants"
	"github.com/kiegroup/kie-cloud-operator/internal/pkg/defaults"
	opv1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/kiegroup/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConstructConsoleObject(t *testing.T) {
	name := "test"
	envReplace := corev1.EnvVar{
		Name:  "KIE_ADMIN_PWD",
		Value: "RedHat",
	}
	envAddition := corev1.EnvVar{
		Name:  "CONSOLE_TEST",
		Value: "test",
	}
	cr := &opv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: opv1.AppSpec{
			Environment: "trial",
			Version:     "7.0",
			Objects: opv1.AppObjects{
				Console: opv1.AppObject{
					Env: []corev1.EnvVar{
						envReplace,
						envAddition,
					},
				},
			},
		},
	}
	env, _, err := defaults.GetEnvironment(cr)
	assert.Nil(t, err)

	object := ConstructObject(env.Console, cr)
	assert.Equal(t, fmt.Sprintf("%s-rhpamcentr", name), object.DeploymentConfigs[0].Name)
	re := regexp.MustCompile("[0-9]+")
	assert.Equal(t, fmt.Sprintf("rhpam%s-businesscentral-openshift:%s", strings.Join(re.FindAllString(cr.Spec.Version, -1), ""), constants.ImageStreamTag), env.Console.DeploymentConfigs[0].Spec.Triggers[0].ImageChangeParams.From.Name)
	assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envReplace, "Environment overriding not functional")
	assert.Contains(t, object.DeploymentConfigs[0].Spec.Template.Spec.Containers[0].Env, envAddition, "Environment additions not functional")
}
