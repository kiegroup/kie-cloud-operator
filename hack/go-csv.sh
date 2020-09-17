#!/bin/sh

source ./hack/go-mod-env.sh

go run ./tools/csv-gen/csv-gen.go
operator-sdk bundle validate deploy/olm-catalog/dev/$(go run getversion.go)
operator-sdk bundle validate deploy/olm-catalog/prod/$(go run getversion.go)