#!/bin/sh

REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAG=0.1

go generate ./...
if [[ -z ${CI} ]]; then
    source hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        mkdir -p target/image
        RESULT=$(md5sum build/_output/bin/kie-cloud-operator)
        MD5=$(echo ${RESULT} | awk {'print $1'})
        cekit-cache -v add --md5 ${RESULT}
        cekit build --redhat --build-engine=osbs \
            --overrides "{'version': '${TAG}', 'artifacts': [{'name': 'kie-cloud-operator', 'md5': '${MD5}'}]}"
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
