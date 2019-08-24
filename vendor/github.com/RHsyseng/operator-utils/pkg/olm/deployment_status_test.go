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

func TestSingleDaemonSetStatus(t *testing.T) {
	obj := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ReadyDeployment",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			NumberReady:            3,
		},
	}
	status := GetSingleDaemonSetStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 3, "Expected three ready deployments")
	assert.Equal(t, "ReadyDeployment-1", status.Ready[0])
	assert.Equal(t, "ReadyDeployment-2", status.Ready[1])
	assert.Equal(t, "ReadyDeployment-3", status.Ready[2])
}

func TestStartingSingleDaemonSetStatus(t *testing.T) {
	obj := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "StartingDeployment",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			NumberReady:            1,
		},
	}
	status := GetSingleDaemonSetStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "StartingDeployment-1", status.Ready[0])
	assert.Len(t, status.Starting, 2, "Expected two starting deployments")
	assert.Equal(t, "StartingDeployment-2", status.Starting[0])
	assert.Equal(t, "StartingDeployment-3", status.Starting[1])
}

func TestStoppedSingleDaemonSetStatus(t *testing.T) {
	obj := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "StoppedDeployment",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 0,
			NumberReady:            0,
		},
	}
	status := GetSingleDaemonSetStatus(obj)
	assert.Len(t, status.Stopped, 1, "Expected one stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 0, "Expected no ready deployments")
}

func TestSingleDeploymentStatus(t *testing.T) {
	three := int32(3)
	obj := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ReadyDeployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &three,
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      3,
			ReadyReplicas: 3,
		},
	}
	status := GetSingleDeploymentStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 3, "Expected three ready deployments")
	assert.Equal(t, "ReadyDeployment-1", status.Ready[0])
	assert.Equal(t, "ReadyDeployment-2", status.Ready[1])
	assert.Equal(t, "ReadyDeployment-3", status.Ready[2])
}

func TestStartingSingleDeploymentStatus(t *testing.T) {
	three := int32(3)
	obj := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "StartingDeployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &three,
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      3,
			ReadyReplicas: 1,
		},
	}
	status := GetSingleDeploymentStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "StartingDeployment-1", status.Ready[0])
	assert.Len(t, status.Starting, 2, "Expected two starting deployments")
	assert.Equal(t, "StartingDeployment-2", status.Starting[0])
	assert.Equal(t, "StartingDeployment-3", status.Starting[1])
}

func TestStoppedSingleDeploymentStatus(t *testing.T) {
	zero := int32(0)
	obj := appsv1.Deployment{
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
	}
	status := GetSingleDeploymentStatus(obj)
	assert.Len(t, status.Stopped, 1, "Expected one stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 0, "Expected no ready deployments")
}

func TestSingleStatefulSetStatus(t *testing.T) {
	three := int32(3)
	obj := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ReadyDeployment",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &three,
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:      3,
			ReadyReplicas: 3,
		},
	}
	status := GetSingleStatefulSetStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 3, "Expected three ready deployments")
	assert.Equal(t, "ReadyDeployment-1", status.Ready[0])
	assert.Equal(t, "ReadyDeployment-2", status.Ready[1])
	assert.Equal(t, "ReadyDeployment-3", status.Ready[2])
}

func TestStartingSingleStatefulSetStatus(t *testing.T) {
	three := int32(3)
	obj := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "StartingDeployment",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &three,
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:      3,
			ReadyReplicas: 1,
		},
	}
	status := GetSingleStatefulSetStatus(obj)
	assert.Len(t, status.Stopped, 0, "Expected no stopped deployments")
	assert.Len(t, status.Ready, 1, "Expected one ready deployment")
	assert.Equal(t, "StartingDeployment-1", status.Ready[0])
	assert.Len(t, status.Starting, 2, "Expected two starting deployments")
	assert.Equal(t, "StartingDeployment-2", status.Starting[0])
	assert.Equal(t, "StartingDeployment-3", status.Starting[1])
}

func TestStoppedSingleStatefulSetStatus(t *testing.T) {
	zero := int32(0)
	obj := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "StoppedDeployment",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &zero,
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:      0,
			ReadyReplicas: 0,
		},
	}
	status := GetSingleStatefulSetStatus(obj)
	assert.Len(t, status.Stopped, 1, "Expected one stopped deployment")
	assert.Equal(t, "StoppedDeployment", status.Stopped[0])
	assert.Len(t, status.Starting, 0, "Expected no starting deployments")
	assert.Len(t, status.Ready, 0, "Expected no ready deployments")
}
