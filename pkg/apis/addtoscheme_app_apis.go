package apis

import (
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	consolev1 "github.com/openshift/api/console/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		api.SchemeBuilder.AddToScheme,
		rbacv1.AddToScheme,
		oappsv1.Install,
		routev1.Install,
		oimagev1.Install,
		buildv1.Install,
		operatorsv1alpha1.AddToScheme,
		monv1.AddToScheme,
		consolev1.Install,
		metav1.AddMetaToScheme,
	)
}
