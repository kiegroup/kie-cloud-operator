module github.com/kiegroup/kie-cloud-operator

go 1.12

require (
	github.com/RHsyseng/console-cr-form v0.0.0-00010101000000-000000000000
	github.com/RHsyseng/operator-utils v0.0.0-00010101000000-000000000000
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/prometheus-operator v0.29.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/google/go-cmp v0.3.1
	github.com/gophercloud/gophercloud v0.6.0 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/openshift/api v3.9.1-0.20190424152011-77b8897ec79a+incompatible
	github.com/openshift/client-go v0.0.0-20190401163519-84c2b942258a
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	github.com/operator-framework/operator-sdk v0.0.0-00010101000000-000000000000
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.1
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0
)

// Pinned to kubernetes-1.14.1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.31.1
	// Pinned to v2.10.0 (kubernetes-1.14.1) so https://proxy.golang.org can
	// resolve it correctly.
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20190525122359-d20e84d0fb64
)

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.11.0

replace github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.2

replace (
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190814194116-a94e914914f4+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190813201236-5a5508328169
)

replace (
	github.com/RHsyseng/console-cr-form => github.com/RHsyseng/console-cr-form v0.0.0-20191121011138-06a2ea30497b
	github.com/RHsyseng/operator-utils => github.com/RHsyseng/operator-utils v0.0.0-20191024171429-aada4a416142
)
