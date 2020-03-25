package kieapp

import (
	"context"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/test"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

func TestYamlSampleCreation(t *testing.T) {
	crNamespacedName := getNamespacedName("namespace", "cr")
	cr := getInstance(crNamespacedName)
	cr.Spec = api.KieAppSpec{
		Environment: api.RhpamTrial,
	}
	service := test.MockService()
	err := service.Create(context.TODO(), cr)
	assert.Nil(t, err)
	reconciler := &Reconciler{Service: service}
	result, err := reconciler.Reconcile(reconcile.Request{NamespacedName: crNamespacedName})
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Duration(500) * time.Millisecond}, result, "Routes should be created, requeued for hostname detection before other resources are created")

	CreateConsoleYAMLSamples(reconciler)
}
