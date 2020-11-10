#!/bin/sh

source ./hack/go-mod-env.sh
VERSION=$(go run getversion.go)

#go run ./tools/csv-gen/csv-gen.go
go run ./tools/csv-gen/csv-gen.go -replaces 7.9.0-2

operator-sdk bundle validate deploy/olm-catalog/dev/${VERSION}
operator-sdk bundle validate deploy/olm-catalog/prod/${VERSION}