# Requirements
 - go v1.10.x
 - operator-sdk v0.1.1
 - dep v0.5.x

# Build
```shell
make
```

# Upload to a container registry -
e.g.
```shell
docker push quay.io/kiegroup/kie-cloud-operator:latest
```

# Deploy to OpenShift
Globally and only once for the whole cluster:
```shell
oc create -f deploy/crds/kieapp_crd.yaml
```

In a project:
```shell
oc create -f deploy/service_account.yaml
oc create -f deploy/role.yaml
oc create -f deploy/role_binding.yaml
oc create -f deploy/operator.yaml
```

# Trigger a KieApp deployment
```shell
oc create -f deploy/crs/kieapp_trial.yaml
```

# Clean up a KieApp deployment:
```shell
oc delete KieApp trial
```

# Development

Change log level at runtime w/ the `LOG_LEVEL` environment variable. e.g. -

```shell
make dep
make clean
LOG_LEVEL="debug" operator-sdk up local --namespace=<namespace>
```
Also at runtime, change registry for rhpam ImageStreamTags -
```shell
INSECURE=true REGISTRY=<registry url> operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to generate, vet, format, and test your code. This all can be done with one command.
```shell
make test
```
