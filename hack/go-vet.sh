#!/bin/sh

source ./hack/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./hack/go-mod.sh
    ./hack/go-sdk-gen.sh
fi
go vet ./...