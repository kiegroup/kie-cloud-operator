#!/bin/sh

source ./hack/go-mod-env.sh

# controller-gen replaces the deprecated `operator-sdk generate k8s` and
# `operator-sdk generate crds` commands from operator-sdk v0.x which are
# incompatible with Go 1.18+ (generics in stdlib break the old code parser).
#
# Install: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0

# Resolve controller-gen: check GOBIN first (where go install puts it), then GOPATH/bin, then PATH
GOBIN_DIR=$(go env GOBIN 2>/dev/null)
GOBIN_DIR="${GOBIN_DIR:-$(go env GOPATH)/bin}"
CONTROLLER_GEN="${GOBIN_DIR}/controller-gen"

if [ ! -x "${CONTROLLER_GEN}" ]; then
    echo "controller-gen not found. Installing..."
    # Unset GOFLAGS temporarily so go install can fetch the module (not subject to -mod=vendor)
    GOFLAGS="" go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0
fi

# Generate deepcopy functions (replaces: operator-sdk generate k8s)
${CONTROLLER_GEN} object paths=./pkg/apis/...

# Generate CRD manifests (replaces: operator-sdk generate crds)
${CONTROLLER_GEN} crd paths=./pkg/apis/... output:crd:artifacts:config=deploy/crds

mv deploy/crds/app.kiegroup.org_kieapps.yaml deploy/crds/kieapp.crd.yaml 2>/dev/null || true

CSVVERSION=$(go run getversion.go -csv)
OLMPATH="deploy/olm-catalog"
for OLMENV in dev test prod
do
    mkdir -p ${OLMPATH}/${OLMENV}/${CSVVERSION}/manifests
    cp -p deploy/crds/kieapp.crd.yaml ${OLMPATH}/${OLMENV}/${CSVVERSION}/manifests/
done
