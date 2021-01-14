#!/bin/sh

source ./hack/go-mod-env.sh

echo
echo Building operator bundle image:
echo

BUNDLE=rhpam-operator-bundle
ENVDIR=olm-catalog/test
BUNDLE_NAME=rhpam-7/${BUNDLE}
VERSION=$(go run getversion.go)
CSV=businessautomation-operator.${VERSION}.clusterserviceversion.yaml
if [[ ${DEV} == true ]]; then
    ENVDIR=olm-catalog/dev
    BUNDLE_NAME=quay.io/tchughesiv/${BUNDLE}
fi
CFLAGS="docker --no-squash"
if [[ ${LOCAL} != true ]]; then
    CFLAGS="osbs"
    if [[ ${1} == "release" ]]; then
        CFLAGS+=" --release"
        ENVDIR=olm-catalog/prod
    fi
fi

echo ${CFLAGS}

VERSION=$(go run getversion.go)
CSVVERSION=$(go run getversion.go -csv)

OLMDIR=deploy/olm-catalog/prod
BUNDLE_NAME=rhpam-7/${BUNDLE}
CSV=businessautomation-operator.${CSVVERSION}.clusterserviceversion.yaml
if [[ ${DEV} == true ]]; then
    OLMDIR=deploy/olm-catalog/dev
    BUNDLE_NAME=quay.io/tchughesiv/${BUNDLE}
fi
VERDIR=${OLMDIR}/${CSVVERSION}
CRD=kieapp.crd.yaml
if (( $(echo "${VERSION} 7.9.0" | awk '{print ($1 < $2)}') )); then
    CRD=kieapp.crd.v1beta1.yaml
fi
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
    --overrides '{name: '${BUNDLE_NAME}'}' \
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
        {
        "container":
            {
            "operator_manifests":
                {"enable_digest_pinning": false, "enable_repo_replacements": false, "enable_registry_replacements": false, "manifests_dir": 'modules/olm-catalog/${CSVVERSION}/manifests'},
            "platforms":
                {"only": ["x86_64"]}
            }
        },
    "repository":
        {"name": "containers/rhpam-operator-bundle", "branch": "rhba-stable-rhel-8"}
    }
}' \
    ${CFLAGS}
