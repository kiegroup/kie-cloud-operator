#!/bin/sh

source ./hack/go-mod-env.sh

operator-sdk generate k8s
operator-sdk generate crds
mv deploy/crds/app.kiegroup.org_kieapps_crd.yaml deploy/crds/kieapp.crd.yaml