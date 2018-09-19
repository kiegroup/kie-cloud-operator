# Build
dep ensure
operator-sdk build bmozaffa/rhpam-operator

# Upload to dockerhub
docker push bmozaffa/rhpam-operator:latest

# Deploy to OpenShift
Globally and only once for the whole cluster:
oc create -f deploy/rbac.yaml
oc create -f deploy/crd.yaml

In a project:
oc create -f deploy/operator.yaml

# Trigger application deployment
oc create -f deploy/trial-environment.yaml

# Clean up in the project:
oc delete all --all
oc delete -f deploy/trial-environment.yaml