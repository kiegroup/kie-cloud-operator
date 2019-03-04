# Kie Cloud Operator

[![Go Report](https://goreportcard.com/badge/github.com/kiegroup/kie-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kie-cloud-operator)

## Requirements

- go v1.10+
- dep v0.5.0+
- operator-sdk v0.3.0

## Build

```bash
make
```

## Upload to a container registry

e.g.

```bash
docker push quay.io/kiegroup/kie-cloud-operator:<version>
```

## Deploy to OpenShift using OLM

As cluster-admin and an OCP 3.11+ cluster with OLM installed, issue the following command:

```bash
# If using the default OLM namespace "operator-lifecycle-manager"
./hack/catalog.sh

# If using a different namespace for OLM
./hack/catalog.sh <namespace>
```

This will create a new `CatalogSource` and `ConfigMap`, allowing the OLM Catalog to see this Operator's `ClusterServiceVersion`.

## Deploy to OpenShift Manually

Globally and only once for the whole cluster:

```bash
oc create -f deploy/crds/kieapp.crd.yaml
```

In a project:

```bash
oc create -f deploy/service_account.yaml
oc create -f deploy/role.yaml
oc create -f deploy/role_binding.yaml
oc create -f deploy/operator.yaml
```

### Trigger a KieApp deployment

```bash
$ oc create -f deploy/crs/kieapp_rhpam_trial.yaml
kieapp.app.kiegroup.org/rhpam-trial created
```

### Clean up a KieApp deployment

```bash
# Using the KieApp name
oc delete kieapp rhpam-trial
# OR using the file name
oc delete -f deploy/crs/kieapp_rhpam_trial.yaml
# OR delete all the KieApp deployments
oc delete kieapp --all
```

## Development

Change log level at runtime w/ the `DEBUG` environment variable. e.g. -

```bash
make dep
make clean
DEBUG="true" operator-sdk up local --namespace=<namespace>
```

Also at runtime, change registry for rhpam ImageStreamTags -

```bash
INSECURE=true REGISTRY=<registry url> operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to generate, vet, format, and test your code. This all can be done with one command.

```bash
make test
```

## Authentication configuration

It is possible to configure RHPAM authentication with an external Identity Provider such as RH-SSO or LDAP.

### SSO

In order to integrate RHPAM authentication with an existing instance of RH-SSO an `auth` element must be provided with a valid `sso` configuration. If the `hostnameHTTPS` is not provided for some client it will be retrieved from the generated route hostname. It is important to say that the URL and Realm parameters are mandatory.

```yaml
spec:
  environment: production
  auth:
    sso:
      url: https://rh-sso.example.com
      realm: rhpam
      adminUser: admin
      adminPassword: secret
      clients:
        console:
          name: rhpamcentr-client
          secret: somePwd
        servers:
          - name: kieserver-client-a
            secret: otherPwd
            hostnameHTTPS: kieserver-a.example.com
          - name: kieserver-client-b
            secret: yetOtherPwd
            hostnameHTTPS: kieserver-b.example.com
```

### LDAP

The LDAP configuration allows RHPAM to authenticate and retrieve the user's groups from an existing LDAP instance. Only the URL parameter is mandatory

```yaml
spec:
  environment: production
  auth:
    ldap:
      url: ldaps://myldap.example.com
      bindDN: uid=admin,dc=example,dc=com
      bindPassword: somePwd
      baseCtxDN: ou=users,dc=example,dc=com
```

### RoleMapper

Finally, it is also possible to provide a properties file including how the roles returned by the external IdP are going to be mapped into application roles.

```yaml
spec:
  environment: production
  auth:
    ldap:
      url: ldaps://myldap.example.com
      bindDN: uid=admin,dc=example,dc=com
      bindPassword: somePwd
      baseCtxDN: ou=users,dc=example,dc=com
    roleMapper:
      rolesProperties: rolesMapper.properties
      replaceRole: true
```

## Build rhel-based image for release

Requires `cekit` and `rhpkg` -
```bash
# local build
make rhel
# scratch build
make rhel-scratch
# release candidate
make rhel-release
```
