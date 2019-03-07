#!/bin/sh

if [[ -z ${CI} ]]; then
    source hack/go-vet.sh
    source hack/go-fmt.sh
    source hack/catalog-source.sh
fi
GOCACHE=off go test ./...