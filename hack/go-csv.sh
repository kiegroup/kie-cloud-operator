#!/bin/sh

source ./hack/go-mod-env.sh
VERSION=$(go run getversion.go -csv)

go run ./tools/csv-gen/csv-gen.go

# operator-sdk bundle validate requires operator-sdk v1.x
# Install: https://sdk.operatorframework.io/docs/installation/
# or: curl -sL https://github.com/operator-framework/operator-sdk/releases/download/v1.38.0/operator-sdk_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m) -o /usr/local/bin/operator-sdk && chmod +x /usr/local/bin/operator-sdk
OPERATOR_SDK_MIN_VERSION="v1"
OPERATOR_SDK_VERSION=$(operator-sdk version 2>/dev/null | grep -o '"v[0-9]*\.[0-9]*\.[0-9]*"' | tr -d '"' || echo "")
if [ -z "${OPERATOR_SDK_VERSION}" ] || [[ "${OPERATOR_SDK_VERSION}" != ${OPERATOR_SDK_MIN_VERSION}* ]]; then
    echo "WARNING: operator-sdk v1.x not found (found: ${OPERATOR_SDK_VERSION:-none}). Skipping bundle validate."
    echo "Install operator-sdk v1.x from: https://sdk.operatorframework.io/docs/installation/"
    exit 0
fi

operator-sdk bundle validate deploy/olm-catalog/dev/${VERSION}
operator-sdk bundle validate deploy/olm-catalog/test/${VERSION}
operator-sdk bundle validate deploy/olm-catalog/prod/${VERSION}
