# Kie Cloud Operator

[![Go Report](https://goreportcard.com/badge/github.com/kiegroup/kie-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kie-cloud-operator)

## Requirements

- go v1.13.x
- operator-sdk v0.19.2

## Build

```bash
make
```

## Upload to a container registry

e.g.

```bash
docker push quay.io/kiegroup/kie-cloud-operator:<version>
```

## Deploy to OpenShift 4.5+ using OLM

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

If pushing to another quay repository, replace _kiegroup_ with your username or other namespace. Also note that the push command does not overwrite an existing repository, and it needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

**Create your own index image**
```bash
$ make bundle-dev
USERNAME=tchughesiv
VERSION=$(go run getversion.go)
IMAGE=quay.io/${USERNAME}/rhpam-operator-bundle
BUNDLE=${IMAGE}:${VERSION}

$ docker push ${BUNDLE}
BUNDLE_DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' ${BUNDLE})
INDEX_VERSION=v4.5 # v4.6
INDEX_IMAGE=quay.io/${USERNAME}/ba-operator-index:${INDEX_VERSION}
INDEX_FROM=${INDEX_IMAGE}_$(go run getversion.go --prior)
INDEX_TO=${INDEX_IMAGE}_${VERSION}

$ opm index add -c docker --bundles ${BUNDLE_DIGEST} --from-index ${INDEX_FROM} --tag ${INDEX_TO}

$ docker push ${INDEX_TO}

# only run in dev env
$ oc patch operatorhub.config.openshift.io/cluster -p='{"spec":{"disableAllDefaultSources":true}}' --type=merge
$ oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-catalog
  namespace: openshift-marketplace
spec:
  displayName: "Dev Bundles"
  publisher: "Red Hat"
  sourceType: grpc
  image: ${INDEX_TO}
  updateStrategy:
    registryPoll:
      interval: 45m
EOF
```

**CRC setup for STAGE testing**
```bash
INDEX_VERSION=v4.5 # v4.6
REGISTRY=registry-proxy.engineering.redhat.com
INDEX_IMAGE=${REGISTRY}/rh-osbs/iib-pub-pending:${INDEX_VERSION}
$ oc patch --type=merge --patch='{
  "spec": {
    "registrySources": {
      "insecureRegistries": [
        "'${REGISTRY}'"
      ]
    }
  }
}' image.config.openshift.io/cluster

$ ssh -i ~/.crc/machines/crc/id_rsa -o StrictHostKeyChecking=no core@$(crc ip) << EOF
  sudo echo " " | sudo tee -a /etc/containers/registries.conf
  sudo echo "[[registry]]" | sudo tee -a /etc/containers/registries.conf
  sudo echo "  location = \"${REGISTRY}\"" | sudo tee -a /etc/containers/registries.conf
  sudo echo "  insecure = true" | sudo tee -a /etc/containers/registries.conf
  sudo echo "  blocked = false" | sudo tee -a /etc/containers/registries.conf
  sudo echo "  mirror-by-digest-only = false" | sudo tee -a /etc/containers/registries.conf
  sudo echo "  prefix = \"\"" | sudo tee -a /etc/containers/registries.conf
  sudo systemctl restart crio
  sudo systemctl restart kubelet
EOF

$ oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-stage-catalog
  namespace: openshift-marketplace
spec:
  displayName: "Stage Bundles"
  publisher: "Red Hat"
  sourceType: grpc
  image: ${INDEX_IMAGE}
  updateStrategy:
    registryPoll:
      interval: 45m
EOF

# $ for TAG in 1.2.0-5 1.2.1-3 1.3.0-3 1.4.0-3 1.4.1-3 7.8.0-3 7.8.1-2; do opm index add -c docker --bundles registry-proxy.engineering.redhat.com/rh-osbs/rhpam-7-rhpam-operator-bundle:${TAG} --from-index ${INDEX_IMAGE} --tag ${INDEX_IMAGE}; docker push ${INDEX_IMAGE}; done
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. It can be easily found by filtering the provider type to _Custom_.

### Trigger a KieApp deployment

Use the OLM console to subscribe to the `Kie Cloud` Operator Catalog Source within your namespace. Once subscribed, use the console to `Create KieApp` or create one manually as seen below.

```bash
$ oc create -f deploy/crs/v2/kieapp_rhpam_trial.yaml
kieapp.app.kiegroup.org/rhpam-trial created
```

### Clean up a KieApp deployment

```bash
oc delete kieapp rhpam-trial
```

## Development

Change log level at runtime w/ the `DEBUG` environment variable. e.g. -

```bash
make mod
make clean
DEBUG="true" operator-sdk run local --watch-namespace<namespace>
```

Also at runtime, change registry for rhpam ImageStreamTags -

```bash
INSECURE=true REGISTRY=<registry url> operator-sdk run local --watch-namespace<namespace>
```

Before submitting PR, please be sure to generate, vet, format, and test your code. This all can be done with one command.

```bash
make test
```

## Build rhel-based image for release

Requires `cekit` v3.7+ and `rhpkg` -

```bash
# local build
make rhel
# scratch build
make rhel-scratch
# release candidate
make rhel-release
```

CSV Generation

```bash
make csv

# OR
# w/ sha lookup/replacement against registry.redhat.io && registry.stage.redhat.io
DIGESTS=true PROD_USER_TOKEN="<username>:<password>" STAGE_USER_TOKEN="<username>:<password>" make csv
```
