#!/bin/sh

source ./hack/go-mod-env.sh

go run ./tools/csv-gen/csv-gen.go
operator-sdk bundle validate deploy/olm-catalog/kiecloud-operator/$(go run getversion.go)
operator-sdk bundle validate deploy/olm-catalog/businessautomation-operator/$(go run getversion.go)