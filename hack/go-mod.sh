#!/bin/sh

source ./hack/go-mod-env.sh

echo Reset vendor directory

if [[ -z ${CI} ]]; then
    go mod tidy
fi

go mod vendor -v