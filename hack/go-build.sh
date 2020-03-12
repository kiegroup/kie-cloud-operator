#!/bin/sh

source ./hack/go-mod-env.sh

REPO=https://github.com/kiegroup/kie-cloud-operator
BRANCH=$(go run getversion.go)
REGISTRY=quay.io/kiegroup
IMAGE=kie-cloud-operator
TAR=${BRANCH}.tar.gz
URL=${REPO}/archive/${TAR}
CFLAGS="docker"

if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
fi

./hack/go-gen.sh

echo
echo Will build console first:
echo
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod=vendor -a -o build/_output/bin/console-cr-form ./cmd/ui
echo

if [[ -z ${CI} ]]; then
    echo Now building operator:
    echo
    operator-sdk build --go-build-args -mod=vendor ${REGISTRY}/${IMAGE}:${BRANCH}
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
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod=vendor -a -o build/_output/bin/kie-cloud-operator ./cmd/manager
fi
