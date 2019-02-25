#!/bin/sh

REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAG=0.1

go generate ./...
if [[ -z ${CI} ]]; then
    source hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        if [[ ${2} == "release" ]]; then
            CFLAG="--build-osbs-release"
        fi
        cekit build ${CFLAG} \
            --redhat \
            --build-tech-preview \
            --package-manager=microdnf \
            --build-engine=osbs \
            --build-osbs-target=rhpam-7-rhel-7-containers-candidate
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
