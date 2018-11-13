# Requirements
 - go v1.10.x
 - operator-sdk v0.1.1

# Build
```shell
dep ensure
go generate ./...
operator-sdk build quay.io/kiegroup/kie-cloud-operator
```

# Upload to a container registry -
e.g.
```shell
docker push quay.io/kiegroup/kie-cloud-operator:latest
```

# Deploy to OpenShift
Globally and only once for the whole cluster:
```shell
oc create -f deploy/crds/app_v1alpha1_kieapp_crd.yaml
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
oc create -f deploy/v1alpha1_crs/kieapp_trial.yaml
```

# Clean up an App deployment:
```shell
oc create -f deploy/v1alpha1_crs/kieapp_trial.yaml
```

# Development

Change log level at runtime w/ the `LOG_LEVEL` environment variable. e.g. -

```shell
dep ensure
LOG_LEVEL="debug" operator-sdk up local --namespace=<namespace>
```

Before submitting PR, please be sure to vet, format, and test your code.
```shell
operator-sdk generate k8s
go vet ./...
go fmt ./...
go test ./...
```
