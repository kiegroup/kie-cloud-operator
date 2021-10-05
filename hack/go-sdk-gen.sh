#!/bin/sh

source ./hack/go-mod-env.sh

operator-sdk generate k8s
operator-sdk generate crds
mv deploy/crds/app.kiegroup.org_kieapps_crd.yaml deploy/crds/kieapp.crd.yaml

CSVVERSION=$(go run getversion.go -csv)
OLMPATH="deploy/olm-catalog"
for OLMENV in dev test prod
do
    mkdir -p ${OLMPATH}/${OLMENV}/${CSVVERSION}/manifests
    cp -p deploy/crds/kieapp.crd.yaml ${OLMPATH}/${OLMENV}/${CSVVERSION}/manifests/
done
