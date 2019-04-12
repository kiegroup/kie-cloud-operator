#!/bin/sh

./hack/catalog-source.sh
oc apply -f deploy/catalog_resources/redhat/catalog-source.yaml