module github.com/kiegroup/kie-cloud-operator

go 1.12

require (
	github.com/RHsyseng/console-cr-form v0.0.0-20200129200812-dc119cb0bd4d
	github.com/RHsyseng/operator-utils v0.0.0-20200204194854-c5b0d8533458
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/prometheus-operator v0.34.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/google/go-cmp v0.3.1
	github.com/gophercloud/gophercloud v0.6.0 // indirect
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
	github.com/imdario/mergo v0.3.8
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20191115003340-16619cd27fa5
	github.com/operator-framework/operator-sdk v0.0.0-00010101000000-000000000000
	github.com/pavel-v-chernykh/keystore-go v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.4.0 // indirect
	github.com/tidwall/sjson v1.0.4
	go.uber.org/zap v1.10.0
	k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver v0.17.2-0.20200115000228-b5a272542936
	k8s.io/apimachinery v0.17.3-beta.0
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	// Pin RHsyseng library versions
	github.com/RHsyseng/console-cr-form => github.com/RHsyseng/console-cr-form v0.0.0-20200129200812-dc119cb0bd4d
	github.com/RHsyseng/operator-utils => github.com/RHsyseng/operator-utils v0.0.0-20200213133748-8821ad0bd623

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	github.com/gobuffalo/packr/v2 => github.com/gobuffalo/packr/v2 v2.7.1

	// Versions after v0.3.7 change behaviour
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.7

	//Openshift release-4.3
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20200205145930-e9d93e317dd1+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20191125132246-f6563a70e19a

	//Operator Framework v0.14.1
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200204235824-a0b8e7fd7a0e
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.14.1

	// Pinned to kubernetes-1.16.2
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9

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
