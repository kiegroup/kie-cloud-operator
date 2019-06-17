package olm

import (
	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDaemonSetStatus(t *testing.T) {
	objs := []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StoppedDeployment",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 0,
				NumberReady:            0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StartingDeployment",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 3,
				NumberReady:            1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ReadyDeployment",
			},
			Status: appsv1.DaemonSetStatus{
				DesiredNumberScheduled: 3,
				NumberReady:            3,
			},
		},
	}
	status := GetDaemonSetStatus(objs)
	assert.Len(t, status.Stopped, 1, "Expected one stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Len(t, status.Starting, 1, "Expected one starting deployment")
	assert.Equal(t, "StartingDeployment", status.Starting[0])
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "ReadyDeployment", status.Ready[0])
}

func TestDeploymentsStatus(t *testing.T) {
	zero := int32(0)
	three := int32(3)
	objs := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StoppedDeployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &zero,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      0,
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StoppedDeployment2",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &three,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      0,
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StartingDeployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &three,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      three,
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ReadyDeployment",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &three,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      three,
				ReadyReplicas: three,
			},
		},
	}
	status := GetDeploymentStatus(objs)
	assert.Len(t, status.Stopped, 2, "Expected two stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Equal(t, "StoppedDeployment2", status.Stopped[1])
	assert.Len(t, status.Starting, 1, "Expected one starting deployment")
	assert.Equal(t, "StartingDeployment", status.Starting[0])
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "ReadyDeployment", status.Ready[0])
}

func TestDeploymentConfigsStatus(t *testing.T) {
	objs := []oappsv1.DeploymentConfig{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StoppedDeployment",
			},
			Spec: oappsv1.DeploymentConfigSpec{
				Replicas: 0,
			},
			Status: oappsv1.DeploymentConfigStatus{
				Replicas:      0,
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StoppedDeployment2",
			},
			Spec: oappsv1.DeploymentConfigSpec{
				Replicas: 3,
			},
			Status: oappsv1.DeploymentConfigStatus{
				Replicas:      0,
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "StartingDeployment",
			},
			Spec: oappsv1.DeploymentConfigSpec{
				Replicas: 3,
			},
			Status: oappsv1.DeploymentConfigStatus{
				Replicas:      3,
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ReadyDeployment",
			},
			Spec: oappsv1.DeploymentConfigSpec{
				Replicas: 3,
			},
			Status: oappsv1.DeploymentConfigStatus{
				Replicas:      3,
				ReadyReplicas: 3,
			},
		},
	}
	status := GetDeploymentConfigStatus(objs)
	assert.Len(t, status.Stopped, 2, "Expected two stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Equal(t, "StoppedDeployment2", status.Stopped[1])
	assert.Len(t, status.Starting, 1, "Expected one starting deployment")
	assert.Equal(t, "StartingDeployment", status.Starting[0])
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "ReadyDeployment", status.Ready[0])
}
