#!/bin/sh
REPO=https://github.com/kiegroup/kie-cloud-operator
BRANCH=1.1.1
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAG=1.1
TAR=${BRANCH}.tar.gz
URL=${REPO}/archive/${TAR}
CFLAGS="docker"

go generate ./...
if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
    echo
    echo Will build console first:
    echo
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -a -o build/_output/bin/console-cr-form github.com/kiegroup/kie-cloud-operator/cmd/ui
    echo
    echo Now building operator:
    echo
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        if [[ ${LOCAL} != true ]]; then
            CFLAGS="osbs"
            if [[ ${2} == "release" ]]; then
                CFLAGS+=" --release"
            fi
        fi
        wget -q ${URL} -O ${TAR}
        MD5=$(md5sum ${TAR} | awk {'print $1'})
        rm ${TAR}

        echo ${CFLAGS}
        cekit --redhat build \
            --overrides "{'artifacts': [{'name': 'kie-cloud-operator.tar.gz', 'md5': '${MD5}', 'url': '${URL}'}]}" \
            ${CFLAGS}
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -a -o build/_output/bin/console-cr-form github.com/kiegroup/kie-cloud-operator/cmd/ui
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
