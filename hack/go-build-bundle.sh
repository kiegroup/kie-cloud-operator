#!/bin/sh

source ./hack/go-mod-env.sh

echo
echo Building operator bundle image:
echo

VERSION=7.8.1
# VERSION=$(go run getversion.go)

if [[ ${UPSTREAM} != true ]]; then
    CFLAGS="docker --no-squash"
    if [[ ${LOCAL} != true ]]; then
        CFLAGS="osbs"
        if [[ ${1} == "release" ]]; then
            CFLAGS+=" --release"
        fi
    fi
    echo ${CFLAGS}

    VERDIR=deploy/olm-catalog/businessautomation-operator/${VERSION}

    CRD=kieapp.crd.yaml
    if (( $(echo "${VERSION} 7.9.0" | awk '{print ($1 < $2)}') )); then
        CRD=kieapp.crd.v1beta1.yaml
    fi
    CSV=businessautomation-operator.${VERSION}.clusterserviceversion.yaml
    ANNO=annotations.yaml
    CSV_PATH=${VERDIR}/manifests/${CSV}
    CRD_PATH=${VERDIR}/manifests/${CRD}
    ANNO_PATH=${VERDIR}/metadata/${ANNO}

    if [[ "$OSTYPE" == "darwin"* ]]; then
        MD5_CSV=$(md5 -q ${CSV_PATH})
        MD5_CRD=$(md5 -q ${CRD_PATH})
        MD5_ANNO=$(md5 -q ${ANNO_PATH})
    else
        MD5_CSV=$(md5sum ${CSV_PATH} | awk '{ printf $1 }')
        MD5_CRD=$(md5sum ${CRD_PATH} | awk '{ printf $1 }')
        MD5_ANNO=$(md5sum ${ANNO_PATH} | awk '{ printf $1 }')
    fi
    cekit-cache add --md5 ${MD5_CSV} ${CSV_PATH}
    cekit-cache add --md5 ${MD5_CRD} ${CRD_PATH}
    cekit-cache add --md5 ${MD5_ANNO} ${ANNO_PATH}

    cekit -v --descriptor image-bundle.yaml --redhat build \
        --overrides '{version: '${VERSION}'}' \
        --overrides '{
    artifacts: [
        {name: '${CSV}', path: '${CSV_PATH}', md5: '${MD5_CSV}', dest: '/manifests'},
        {name: '${CRD}', path: '${CRD_PATH}', md5: '${MD5_CRD}', dest: '/manifests'},
        {name: '${ANNO}', path: '${ANNO_PATH}', md5: '${MD5_ANNO}', dest: '/metadata'}
    ]}' \
        --overrides '{
    "osbs":
        {
        "configuration":
            {"container":
                {
                "operator_manifests":
                    {"enable_digest_pinning": false, "enable_repo_replacements": false, "enable_registry_replacements": false, "manifests_dir": 'modules/olm-catalog/${VERSION}/manifests'},
                "platforms":
                    {"only": ["x86_64"]}
                }
            },
        "repository":
            {"name": "containers/rhpam-operator-bundle", "branch": "rhba-stable-rhel-8"}
        }
    }' \
        ${CFLAGS}
    exit 0
fi

docker build -f build/bundle.Dockerfile deploy/olm-catalog/kiecloud-operator/${VERSION} -t quay.io/kiegroup/kiecloud-operator-bundle:${VERSION}