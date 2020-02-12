package apis

import (
	v1 "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	api "github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v2"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		api.SchemeBuilder.AddToScheme,
		v1.SchemeBuilder.AddToScheme,
		rbacv1.AddToScheme,
		oappsv1.Install,
		routev1.Install,
		oimagev1.Install,
		buildv1.Install,
		operatorsv1alpha1.AddToScheme,
	)
}
