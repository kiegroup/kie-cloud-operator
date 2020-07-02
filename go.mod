module github.com/kiegroup/kie-cloud-operator

go 1.13

require (
	github.com/RHsyseng/console-cr-form v0.0.0-20200414161125-135bc9b52976
	github.com/RHsyseng/operator-utils v0.0.0-20200811204138-48b5b595439a
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/prometheus-operator v0.41.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-openapi/spec v0.19.9
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/google/go-cmp v0.5.0
	github.com/google/uuid v1.1.1
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
<<<<<<< HEAD
	github.com/imdario/mergo v0.3.9
	github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
=======
	github.com/imdario/mergo v0.3.8
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
>>>>>>> 0acab22e... golang 1.13
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/operator-framework/api v0.3.8
	github.com/operator-framework/operator-lifecycle-manager v3.11.0+incompatible
	github.com/operator-framework/operator-sdk v0.19.2
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.4.0
	github.com/tidwall/sjson v1.0.4
	golang.org/x/mod v0.2.0
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

	// Pin RHsyseng library versions
	github.com/RHsyseng/console-cr-form => github.com/RHsyseng/console-cr-form v0.0.0-20200414161125-135bc9b52976
	github.com/RHsyseng/operator-utils => github.com/RHsyseng/operator-utils v0.0.0-20200811204138-48b5b595439a

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.7.1

	// Versions after v0.3.7 change behaviour
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.7

	// Openshift release-4.5
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c

	// Operator Framework v0.19.2
	github.com/operator-framework/api => github.com/operator-framework/api v0.3.11
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.19.2

	// Pinned to kubernetes-1.18.2
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
	k8s.io/kubernetes => k8s.io/kubernetes v0.18.2

	// others
	modernc.org/cc => gitlab.com/cznic/cc v1.0.0
	modernc.org/golex => gitlab.com/cznic/golex v1.0.0
	modernc.org/mathutil => gitlab.com/cznic/mathutil v1.0.0
	modernc.org/strutil => gitlab.com/cznic/strutil v1.0.0
	modernc.org/xc => gitlab.com/cznic/xc v1.0.0
	mvdan.cc/interfacer => github.com/mvdan/interfacer v0.0.0-20180901003855-c20040233aed
	mvdan.cc/lint => github.com/mvdan/lint v0.0.0-20170908181259-adc824a0674b
	mvdan.cc/unparam => github.com/mvdan/unparam v0.0.0-20190209190245-fbb59629db34
)
