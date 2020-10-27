#!/bin/sh

source ./hack/go-mod-env.sh

echo Reset vendor directory

go mod tidy
go mod vendor

if [[ -n ${CI} ]]; then
    git diff --exit-code
fi
