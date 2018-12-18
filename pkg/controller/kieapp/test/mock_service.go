package test

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	"github.com/kiegroup/kie-cloud-operator/pkg/controller/kieapp/logs"
	routev1 "github.com/openshift/api/route/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logs.GetLogger("kieapp.test")

type MockPlatformService struct{}

func (MockPlatformService) GetClient() client.Client {
	return fake.NewFakeClient()
}

func (MockPlatformService) GetRouteHost(route routev1.Route, cr *v1.KieApp) string {
	return "www.example.com"
}

func (MockPlatformService) UpdateObj(obj v1.OpenShiftObject) (reconcile.Result, error) {
	log.Debugf("Mock service will do no-op in lieu of updating %v", obj)
	return reconcile.Result{}, nil
}

func (MockPlatformService) CreateCustomObjects(object v1.CustomObject, cr *v1.KieApp) (reconcile.Result, error) {
	log.Debugf("Mock service will do no-op in lieu of creating %v with CR %v", object, cr)
	return reconcile.Result{}, nil
}
