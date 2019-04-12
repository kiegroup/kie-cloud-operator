#!/bin/sh

if [[ -z ${1} ]]; then
    CATALOG_NS="operator-lifecycle-manager"
else
    CATALOG_NS=${1}
fi

CSV=`cat deploy/catalog_resources/redhat/businessautomation-operator.1.0.1.clusterserviceversion.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
CRD=`cat deploy/crds/kieapp.crd.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
PKG=`cat deploy/catalog_resources/redhat/businessautomation.1.0.1.package.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`

cat << EOF > deploy/catalog_resources/redhat/catalog-source.yaml
apiVersion: v1
kind: List
items:
  - apiVersion: v1
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

  - apiVersion: operators.coreos.com/v1alpha1
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
