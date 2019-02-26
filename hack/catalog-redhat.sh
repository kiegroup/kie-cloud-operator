#!/bin/sh

if [[ -z ${1} ]]; then
    CATALOG_NS="operator-lifecycle-manager"
else
    CATALOG_NS=${1}
fi

CSV=`cat deploy/catalog_resources/redhat/businessautomation-operator.v1.0.0.clusterserviceversion.yaml | sed -e 's/^/      /' | sed '0,/ /{s/      /    - /}'`
CRD=`cat deploy/crds/kieapp.crd.yaml | sed -e 's/^/      /' | sed '0,/ /{s/      /    - /}'`
PKG=`cat deploy/catalog_resources/redhat/businessautomation.package.yaml | sed -e 's/^/      /' | sed '0,/ /{s/      /    - /}'`

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: ba-resources
  namespace: ${CATALOG_NS}
data:
  clusterServiceVersions: |
${CSV}
  customResourceDefinitions: |
${CRD}
  packages: >
${PKG}
EOF

cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ba-resources
  namespace: ${CATALOG_NS}
spec:
  configMap: ba-resources
  displayName: Business Automation Operators
  publisher: Red Hat
  sourceType: internal
status:
  configMapReference:
    name: ba-resources
    namespace: ${CATALOG_NS}
EOF

cat <<EOF >deploy/catalog_resources/redhat/bundle.1.0.0.yaml
data:
  clusterServiceVersions: |
${CSV}
  customResourceDefinitions: |
${CRD}
  packages: >
${PKG}
EOF
