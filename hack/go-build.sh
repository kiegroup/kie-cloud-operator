#!/bin/sh
REPO=https://github.com/kiegroup/kie-cloud-operator
BRANCH=1.1.x
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAG=1.1
TAR=${BRANCH}.tar.gz
URL=${REPO}/archive/${TAR}
CFLAGS="--redhat --build-tech-preview"

go generate ./...
if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
    if [[ ${1} == "rhel" ]]; then
        if [[ ${LOCAL} != true ]]; then
            CFLAGS+=" --build-engine=osbs --build-osbs-target=rhba-7.4-openshift-containers-candidate" # rhpam-7-rhel-7-containers-candidate
            if [[ ${2} == "release" ]]; then
                CFLAGS+=" --build-osbs-release"
            fi
        fi
        wget -q ${URL} -O ${TAR}
        MD5=$(md5sum ${TAR} | awk {'print $1'})
        rm ${TAR}

        echo ${CFLAGS}
        cekit build ${CFLAGS} \
            --overrides "{'artifacts': [{'name': 'kie-cloud-operator.tar.gz', 'md5': '${MD5}', 'url': '${URL}'}]}"
    fi
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o build/_output/bin/kie-cloud-operator github.com/kiegroup/kie-cloud-operator/cmd/manager
fi
