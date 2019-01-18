#!/bin/sh

if [[ -z ${CI} ]]; then
    source hack/go-vet.sh
    source hack/go-fmt.sh
fi
GOCACHE=off go test ./...