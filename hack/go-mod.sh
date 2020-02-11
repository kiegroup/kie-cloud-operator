#!/bin/sh

. ./hack/go-mod-env.sh

echo Reset vendor directory

setGoModEnv

if [[ -z ${CI} ]]; then
    go mod tidy
else
    go mod tidy -v
fi

go mod vendor
