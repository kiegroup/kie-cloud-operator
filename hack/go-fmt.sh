#!/bin/sh

gofmt -s -l -w cmd/ pkg/ version/ tools/

if [[ -n ${CI} ]]; then
    git diff --exit-code
fi
