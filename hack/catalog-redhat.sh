#!/bin/sh

./hack/catalog-source.sh
kubectl apply -f deploy/catalog_resources/redhat/catalog-source.yaml