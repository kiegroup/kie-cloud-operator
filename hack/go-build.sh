#!/bin/sh

source ./hack/go-mod-env.sh

REPO=https://github.com/kiegroup/kie-cloud-operator
PRODUCT_VERSION=$(go run getversion.go)
OPERATOR_VERSION=$(go run getversion.go -csv)
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAR=modules/builder/${IMAGE}.tar.gz
OVERRIDE_IMG_DESCRIPTOR=""

URL=${REPO}/archive/${OPERATOR_VERSION}.tar.gz
if [ -z ${BRANCH_NIGHTLY} ]; then
  BRANCH_NIGHTLY="release-v7.13.x-blue"
fi
URL_NIGHTLY=${REPO}/tarball/${BRANCH_NIGHTLY}

CFLAGS="${1}"

if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
fi

./hack/go-gen.sh

if [[ -z ${CI} || -n ${CEKIT_OSBS_BUILD} ]]; then
    echo Now building operator:
    echo
    if [[ ${2} == "rhel" ]]; then
        if [[ ${LOCAL} != true ]]; then
            CFLAGS="osbs"
            if [[ ${3} == "release" ]]; then
                CFLAGS+=" --release"
                wget -q ${URL} -O ${TAR}
            fi
            if [[ ${3} == "nightly" ]]; then
              OVERRIDE_IMG_DESCRIPTOR=" --descriptor image-prod.yaml"
            fi
            if [[ ! -z ${CEKIT_RESPOND_YES+z} ]]; then
                    CFLAGS+=" -y"
            fi
        fi

        echo ${CFLAGS}
        cekit --verbose --redhat ${OVERRIDE_IMG_DESCRIPTOR} build \
           --overrides '{version: '${PRODUCT_VERSION}'}' \
           ${CFLAGS}
        if [[ -f ${TAR} ]] ; then
          rm ${TAR}
        fi
    else
        echo
        echo Will build console first:
        echo
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod=vendor -a -o build/_output/bin/console-cr-form ./cmd/ui
        echo

        operator-sdk build --go-build-args -mod=vendor ${REGISTRY}/${IMAGE}:${PRODUCT_VERSION}
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod=vendor -a -o build/_output/bin/console-cr-form ./cmd/ui
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod=vendor -a -o build/_output/bin/kie-cloud-operator ./cmd/manager
fi
