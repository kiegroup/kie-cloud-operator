#!/bin/sh

. ./hack/go-mod-env.sh

echo Reset vendor diectory

setGoModEnv

go mod tidy
go mod vendor