#!/bin/sh

source ./hack/go-mod-env.sh

REPO=https://github.com/kiegroup/kie-cloud-operator
PRODUCT_VERSION=$(go run getversion.go)
OPERATOR_VERSION=$(go run getversion.go -csv)
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAR=modules/builder/${IMAGE}.tar.gz

URL=${REPO}/archive/${OPERATOR_VERSION}.tar.gz

CFLAGS="docker"

if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
fi

./hack/go-gen.sh

if [[ -z ${CI} ]]; then
    echo Now building operator:
    echo
    if [[ ${1} == "rhel" ]]; then
        if [[ ${LOCAL} != true ]]; then
            CFLAGS="osbs"
            if [[ ${2} == "release" ]]; then
                CFLAGS+=" --release"
            fi
        fi
        wget -q ${URL} -O ${TAR}
        echo ${CFLAGS}
        cekit --verbose --redhat build \
           --overrides '{version: '${PRODUCT_VERSION}'}' \
           ${CFLAGS}
        rm ${TAR}
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
