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

# Trigger application deployment
```shell
oc create -f deploy/crs/kieapp_trial.yaml
```

# Clean up an App deployment:
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

Before submitting PR, please be sure to vet, generate, format, and test your code. This can all be done with one command.
```shell
make test
```
