#!/bin/sh

. ./hack/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./hack/go-vet.sh
    ./hack/go-fmt.sh
fi

setGoModEnv

go test -mod=vendor -count=1 ./...