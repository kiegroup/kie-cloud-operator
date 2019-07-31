// NOTE: Boilerplate only.  Ignore this file.

// Package v2 contains API Schema definitions for the app v2 API group
// +k8s:deepcopy-gen=package,register
// +groupName=app.kiegroup.org
package v2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "app.kiegroup.org", Version: "v2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
