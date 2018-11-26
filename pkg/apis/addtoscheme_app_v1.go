package apis

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	oappsv1 "github.com/openshift/api/apps/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1.SchemeBuilder.AddToScheme,
		rbacv1.SchemeBuilder.AddToScheme,
		oappsv1.SchemeBuilder.AddToScheme,
		routev1.SchemeBuilder.AddToScheme,
		oimagev1.SchemeBuilder.AddToScheme,
	)
}
