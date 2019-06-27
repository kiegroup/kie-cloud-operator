#!/bin/sh

if [[ -z ${CI} ]]; then
    ./hack/go-vet.sh
    ./hack/go-fmt.sh
fi
go test -mod=vendor -count=1 ./...