#!/bin/sh
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAG=1.0
CFLAGS="--redhat --build-tech-preview"

go generate ./...
if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        if [[ ${LOCAL} != true ]]; then
            CFLAGS+=" --build-engine=osbs --build-osbs-target=rhba-7.3-openshift-containers-candidate" # rhpam-7-rhel-7-containers-candidate
            if [[ ${2} == "release" ]]; then
                CFLAGS+=" --build-osbs-release"
            fi
        fi
        echo ${CFLAGS}
        cekit build ${CFLAGS}
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
