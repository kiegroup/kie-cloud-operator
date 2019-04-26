# Kie Cloud Operator

[![Go Report](https://goreportcard.com/badge/github.com/kiegroup/kie-cloud-operator)](https://goreportcard.com/report/github.com/kiegroup/kie-cloud-operator)

## Requirements

- go v1.11+
- dep v0.5.x
- operator-sdk v0.7.0

## Build

```bash
make
```

## Upload to a container registry

e.g.

```bash
docker push quay.io/kiegroup/kie-cloud-operator:<version>
```

## Deploy to OpenShift 4 using OLM

To install this operator on OpenShift 4 for end-to-end testing, make sure you have access to a quay.io account to create an application repository. Follow the [authentication](https://github.com/operator-framework/operator-courier/#authentication) instructions for Operator Courier to obtain an account token. This token is in the form of "basic XXXXXXXXX" and both words are required for the command.

Push the operator bundle to your quay application repository as follows:

```bash
operator-courier push deploy/catalog_resources/courier/bundle_dir/1.1.0 kiegroup kiecloud-operator 1.1.0 "basic XXXXXXXXX"
# operator-courier push deploy/catalog_resources/courier/bundle_dir/1.0.1 kiegroup kiecloud-operator 1.0.1 "basic XXXXXXXXX"
```

If pushing to another quay repository, replace _kiegroup_ with your username or other namespace. Also note that the push command does not overwrite an existing repository, and it needs to be deleted before a new version can be built and uploaded. Once the bundle has been uploaded, create an [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster) to load your operator bundle in OpenShift.

```bash
oc create -f deploy/catalog_resources/courier/kiecloud-operatorsource.yaml
```

Remember to replace _registryNamespace_ with your quay namespace. The name, display name and publisher of the operator are the only other attributes that may be modified.

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the OpenShift console _Catalog_. It can be easily found by filtering the provider type to _Custom_.

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

Requires `cekit` v3 and `rhpkg` -

```bash
# local build
make rhel
# scratch build
make rhel-scratch
# release candidate
make rhel-release
```
