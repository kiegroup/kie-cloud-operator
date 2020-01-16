#!/bin/sh

if [[ -z ${CI} ]]; then
    ./hack/go-mod.sh
    operator-sdk generate k8s
fi
go vet ./...