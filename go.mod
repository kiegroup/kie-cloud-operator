module github.com/kiegroup/kie-cloud-operator

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.9 // indirect
	github.com/Azure/go-autorest v11.5.2+incompatible // indirect
	github.com/RHsyseng/console-cr-form v0.0.0-20190614162449-0d455ea16817
	github.com/RHsyseng/operator-utils v0.0.0-20190616174704-ab479a3d6e3c
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/coreos/prometheus-operator v0.26.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/validate v0.18.0 // indirect
	github.com/gobuffalo/packr/v2 v2.2.0
	github.com/google/go-cmp v0.2.0
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190318015731-ff9851476e98 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/openshift/api v3.9.1-0.20190424152011-77b8897ec79a+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36
	github.com/operator-framework/operator-sdk v0.8.1-0.20190702165559-d6903be02d4e
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.2.0
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	go.opencensus.io v0.19.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	k8s.io/kube-state-metrics v1.6.0 // indirect
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.10
	sigs.k8s.io/testing_frameworks v0.1.0 // indirect
)

// Pinned to kubernetes-1.13.4
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222131558-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221124651-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.6.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)

replace github.com/openshift/api => github.com/openshift/api v3.9.0+incompatible
