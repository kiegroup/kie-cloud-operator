#!/bin/sh

source ./hack/go-mod-env.sh

go generate -mod=vendor ./...