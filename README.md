# Kie Cloud Operator

[![Go Report](https://goreportcard.com/badge/github.com/kiegroup/kie-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kie-cloud-operator)

## Requirements

- go v1.13.x
- operator-sdk v0.19.2
- docker
- [opm](https://github.com/operator-framework/operator-registry/releases)
- [podman](https://podman.io/)
- [cekit](https://cekit.io/)

## Build

```bash
make
```

## Upload to a container registry

e.g.
```bash
docker push quay.io/kiegroup/kie-cloud-operator:<version>
```

## Deploy to OpenShift 4.7+ using OLM

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io (https://quay.io/) account to create
an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication)
instructions for Operator Courier to obtain an account token.
This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Also note that the push command does not overwrite an existing repository,
and it needs to be deleted before a new version can be built and uploaded.
Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster)
to load your operator bundle in OpenShift.

**Create your own Bundle**

i.e. 7.12.0-1 version
Remove the following line from deploy/olm-catalog/dev/7.12.0-1/manifest/businessautomation-operator.clusterserviceversion.yaml

```console
replaces: businessautomation-operator.<last-version>
```

Activate Cekit and run the following command
```console
$ make bundle-dev
```

the last log line is something like this:
```console
INFO  Image built and available under following tags: quay.io/<your_quay_username>/rhpam-operator-bundle:7.12.1, quay.io/${USERNAME}/rhpam-operator-bundle:latest
```
**Push the bundle on the container registry**

VERSION=$(go run getversion.go)

USERNAME=<your_quay_username>
```console
$ docker push quay.io/${USERNAME}/rhpam-operator-bundle:${VERSION}
```

#### Build the index
```console
opm index add --bundles quay.io/${USERNAME}/rhpam-operator-bundle:${VERSION} --tag quay.io/${USERNAME}/rhpam-operator-index:${VERSION}
```
#### Push the index on the container registry
First time login into your quay account
```console
podman login quay.io
```
Push the index on your quay repository
```console
podman push quay.io/${USERNAME}/rhpam-operator-index:${VERSION}
```

#### Disable default catalog sources on Openshift
To test your Operator, with bundle and index you need to disable the default source like the operator hub
```console
oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": true}]'
```

#### catalog-source

A catalog source is repository of CSVs, CRDs, and packages that define an application.

```console
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: xxxxxname
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: xxxxximage
  displayName: My Operator Catalog
  publisher: grpc
```

Choose a CATALOG_SOURCE_NAME something like "max-operator-manifests"

Example of catalog-source.yaml
```console
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: max-operator-manifests
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: quay.io/<your_quay_id>/rhpam-operator-index:7.12.1
  displayName: My Operator Catalog
  publisher: grpc
```

#### Create catalog source on Openshift
```console
oc create -f catalog-source.yaml
```

#### Subscription

A subscription keeps CSVs up to date by tracking a channel in a package.

```console
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: businessautomation-operator
  namespace: <your-namespace>
spec:
  channel: stable
  name: businessautomation-operator
  source: $CATALOG_SOURCE_NAME
  sourceNamespace: openshift-marketplace
```

Example of subscription.yaml

```console
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: businessautomation-operator
  namespace: max
spec:
  channel: stable
  name: businessautomation-operator
  source: max-operator-manifests
  sourceNamespace: openshift-marketplace
```

#### Create subscription
```console
oc create -f subscription.yaml
```
On OpenShift go to your project (e.g. max) to see your subscription and your operator,
this could take a variable time to be visible.


#### Cleanup catalog-source
After your test are completed, to restore the Operator hub and remove your catalog source
delete your catalog source
and run the following command:
```console
oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": false}]'
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_.
It can be easily found by filtering the provider type to _Custom_.

### Trigger a KieApp deployment

Use the OLM console to subscribe to the `Kie Cloud` Operator Catalog Source within your namespace. Once subscribed,
use the console to `Create KieApp` or create one manually as seen below.

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
DEBUG="true" operator-sdk run local --watch-namespace <namespace>
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

Requires `cekit` v3.11 and `rhpkg` -

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
```
