#!/bin/sh

source ./hack/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./hack/go-vet.sh
    ./hack/go-fmt.sh
fi

go test -count=1 ./...