package apis

import (
	"github.com/kiegroup/kie-cloud-operator/pkg/apis/app/v1"
	oappsv1 "github.com/openshift/api/apps/v1"
	authv1 "github.com/openshift/api/authorization/v1"
	routev1 "github.com/openshift/api/route/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1.SchemeBuilder.AddToScheme,
		rbac_v1.SchemeBuilder.AddToScheme,
		oappsv1.SchemeBuilder.AddToScheme,
		authv1.SchemeBuilder.AddToScheme,
		routev1.SchemeBuilder.AddToScheme,
	)
}
