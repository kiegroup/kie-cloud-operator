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

### Trigger a KieApp deployment

Use the OLM console to subscribe to the `Kie Cloud` Operator Catalog Source within your namespace. Once subscribed, use the console to `Create KieApp` or create one manually as seen below.

```bash
$ oc create -f deploy/crs/kieapp_rhpam_trial.yaml
kieapp.app.kiegroup.org/rhpam-trial created
```

### Clean up a KieApp deployment

```bash
oc delete kieapp rhpam-trial
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
  environment: rhpam-authoring
  auth:
    sso:
      url: https://rh-sso.example.com
      realm: rhpam
      adminUser: admin
      adminPassword: secret
  objects:
    console:
      ssoClient:
        name: rhpam-console
        secret: somePwd
    servers:
      - name: kieserver-one
        deployments: 2
        ssoClient:
          name: kieserver-one
          secret: otherPwd
          hostnameHTTPS: kieserver-one.example.com
      - name: kieserver-two
        ssoClient:
          name: kieserver-two
          secret: yetOtherPwd
```

### LDAP

The LDAP configuration allows RHPAM to authenticate and retrieve the user's groups from an existing LDAP instance. Only the URL parameter is mandatory

```yaml
spec:
  environment: rhpam-production
  auth:
    ldap:
      url: ldaps://myldap.example.com
      bindDN: uid=admin,ou=users,ou=exmample,ou=com
      bindCredential: s3cret
      baseCtxDN: ou=users,ou=example,ou=com
```

### RoleMapper

Finally, it is also possible to provide a properties file including how the roles returned by the external IdP are going to be mapped into application roles.

```yaml
spec:
  environment: rhpam-production
  auth:
    ldap:
      url: ldaps://myldap.example.com
      bindDN: uid=admin,ou=users,ou=exmample,ou=com
      bindCredential: s3cret
      baseCtxDN: ou=users,ou=example,ou=com
    roleMapper:
      rolesProperties: /conf/roleMapper.properties
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
