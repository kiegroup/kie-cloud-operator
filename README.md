# Requirements
 - go v1.10.4
 - operator-sdk v0.0.7

# Build
```shell
dep ensure
go generate ./...
operator-sdk build quay.io/kiegroup/kie-cloud-operator
```

# Upload to a container registry
```shell
docker push quay.io/kiegroup/kie-cloud-operator:latest
```

# Deploy to OpenShift
Globally and only once for the whole cluster:
```shell
oc create -f deploy/crd.yaml
```

In a project:
```shell
oc create -f deploy/rbac.yaml
oc create -f deploy/operator.yaml
```

# Trigger application deployment
```shell
oc create -f deploy/cr-trial.yaml
```

# Clean up an App deployment:
```shell
oc delete -f deploy/cr-trial.yaml
```

# Development

Change log level at runtime w/ the `LOG_LEVEL` environment variable. e.g. -

```shell
dep ensure
LOG_LEVEL="debug" operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to format and test your code.
```shell
go fmt ./...
go test ./...
```
