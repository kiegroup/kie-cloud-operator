#!/bin/sh

REGISTRY=quay.io/kiegroup
RH_REGISTRY=registry.redhat.io/rhpam-7-tech-preview
IMAGE=kie-cloud-operator
TAG=0.1

go generate ./...
if [[ -z ${CI} ]]; then
    source hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        REGISTRY=${RH_REGISTRY}
        echo "Building Docker image ${REGISTRY}/${IMAGE}:${TAG}"
        docker build . -f build/Dockerfile.rhel -t ${REGISTRY}/${IMAGE}:${TAG}
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
