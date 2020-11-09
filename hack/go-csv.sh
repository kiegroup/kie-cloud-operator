#!/bin/sh

source ./hack/go-mod-env.sh
VERSION=$(go run getversion.go)

#go run ./tools/csv-gen/csv-gen.go
go run ./tools/csv-gen/csv-gen.go -version 7.9.0-2 -replaces ${VERSION}

#operator-sdk bundle validate deploy/olm-catalog/dev/${VERSION}
#operator-sdk bundle validate deploy/olm-catalog/prod/${VERSION}
operator-sdk bundle validate deploy/olm-catalog/dev/7.9.0-2
operator-sdk bundle validate deploy/olm-catalog/prod/7.9.0-2