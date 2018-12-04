package kieapp

import (
	"testing"

	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/defaults"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRetrieveNewEnvironments(t *testing.T) {
	envs := []string{"trial", "authoring", "production"}
	for _, envName := range envs {
		env, err := getNewEnvironment(envName)
		assert.Nil(t, err, "Error retrieving new environment: %v", err)
		assert.NotNil(t, env, "Environment %v returned as nil", envName)
		//
		//bytes, err := yaml.Marshal(env)
		//assert.Nil(t, err, "Error marshalling environment %v", env)
		//_, _ = fmt.Printf("Environment %v:\n\n%v\n", envName, string(bytes))
	}
}

func getNewEnvironment(name string) (v1.Environment, error) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: name,
		},
	}

	env, err := defaults.GetLiteEnvironment(cr)
	if err != nil {
		return v1.Environment{}, err
	}
	return env, nil
}

func TestRetrieveOldEnvironments(t *testing.T) {
	envs := []string{"trial", "authoring", "production"}
	for _, envName := range envs {
		env, err := getOldEnvironment(envName)
		assert.Nil(t, err, "Error retrieving old environment: %v", err)
		assert.NotNil(t, env, "Environment %v returned as nil", envName)
		//
		//bytes, err := yaml.Marshal(env)
		//assert.Nil(t, err, "Error marshalling environment %v", env)
		//_, _ = fmt.Printf("Environment %v:\n\n%v\n", envName, string(bytes))
	}
}

func getOldEnvironment(name string) (v1.Environment, error) {
	cr := &v1.KieApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
		Spec: v1.KieAppSpec{
			Environment: name,
		},
	}

	env, common, err := defaults.GetEnvironment(cr, fake.NewFakeClient())
	if err != nil {
		return v1.Environment{}, err
	}

	env = ConsolidateObjects(env, common, cr)
	return env, nil
}

func TestTrial(t *testing.T) {
	newTrial, err := getNewEnvironment("trial")
	assert.Nil(t, err, "Error retrieving new environment: %v", err)
	assert.NotNil(t, newTrial, "Trial environment returned as nil")
	nilOutEmptyEnvironmentArrays(&newTrial)

	oldTrial, err := getOldEnvironment("trial")
	assert.Nil(t, err, "Error retrieving old environment: %v", err)
	assert.NotNil(t, oldTrial, "Trial environment returned as nil")
	nilOutEmptyEnvironmentArrays(&oldTrial)

	//Change back some of the known updates to the new trial env:
	newTrial.Console.Services[0].ObjectMeta.Annotations = nil

	//Intentional differences prevent deep equals test
	//assert.EqualValues(t, oldTrial, newTrial, "Trail environment does not match")
}

func nilOutEmptyEnvironmentArrays(environment *v1.Environment) {
	nilOutEmptyCustomObjectArrays(&environment.Console)
	for index := range environment.Others {
		nilOutEmptyCustomObjectArrays(&environment.Others[index])
	}
	for index := range environment.Servers {
		nilOutEmptyCustomObjectArrays(&environment.Servers[index])
	}
}

func nilOutEmptyCustomObjectArrays(object *v1.CustomObject) {
	if len(object.DeploymentConfigs) == 0 {
		object.DeploymentConfigs = nil
	}
	if len(object.Routes) == 0 {
		object.Routes = nil
	}
	if len(object.Services) == 0 {
		object.Services = nil
	}
	if len(object.RoleBindings) == 0 {
		object.RoleBindings = nil
	}
	if len(object.Secrets) == 0 {
		object.Secrets = nil
	}
	if len(object.ServiceAccounts) == 0 {
		object.ServiceAccounts = nil
	}
	if len(object.PersistentVolumeClaims) == 0 {
		object.PersistentVolumeClaims = nil
	}
}
