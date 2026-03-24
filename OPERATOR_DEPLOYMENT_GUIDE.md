# KIE Cloud Operator - Build, Deploy, and Troubleshooting Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Build Process](#build-process)
3. [Deployment Process](#deployment-process)
4. [Verification Steps](#verification-steps)
5. [Issues and Resolutions](#issues-and-resolutions)
6. [KieApp Deployment](#kieapp-deployment)

---

## Prerequisites

### Required Tools
- Go v1.18.x
- operator-sdk v0.19.2
- Docker or Podman
- opm (Operator Package Manager)
- cekit v3.11+
- OpenShift CLI (oc)
- Access to container registries:
  - quay.io (for operator images)
  - na.artifactory.swg-devops.com (for BAMOE images)

### Environment Setup
```bash
# Set your registry username
export USERNAME=<your-quay-username>

# Get operator version
VERSION=$(go run getversion.go)
echo "Version: $VERSION"
```

---

## Build Process

### 1. Build Operator Image

```bash
# Build the operator
make

# This creates: quay.io/kiegroup/kie-cloud-operator:8.0.9
```

**Output:** Docker image built locally

### 2. Tag and Push Operator Image

```bash
# Tag for your registry
docker tag quay.io/kiegroup/kie-cloud-operator:8.0.9 \
  quay.io/${USERNAME}/kie-cloud-operator:8.0.9

# Push to registry
docker push quay.io/${USERNAME}/kie-cloud-operator:8.0.9
```

### 3. Update CSV with Custom Image

Edit `deploy/olm-catalog/dev/8.0.9-1/manifests/bamoe-businessautomation-operator.clusterserviceversion.yaml`:

```yaml
# Line 238: Change from
image: quay.io/kiegroup/kie-cloud-operator:8.0.9

# To
image: quay.io/<your-username>/kie-cloud-operator:8.0.9
```

### 4. Build Bundle Image

```bash
# Build bundle with updated CSV
make bundle-dev

# This creates: quay.io/${USERNAME}/rhpam-operator-bundle:8.0.9
```

**Output:** Bundle image containing CSV and CRDs

### 5. Push Bundle Image

```bash
podman push quay.io/${USERNAME}/rhpam-operator-bundle:${VERSION}
```

### 6. Build Index Image

```bash
# Build catalog index
opm index add \
  --bundles quay.io/${USERNAME}/rhpam-operator-bundle:${VERSION} \
  --tag quay.io/${USERNAME}/rhpam-operator-index:${VERSION}
```

**Output:** Index image for OLM catalog

### 7. Push Index Image

```bash
podman push quay.io/${USERNAME}/rhpam-operator-index:${VERSION}
```

---

## Deployment Process

### 1. Disable Default Catalog Sources (Optional for Testing)

```bash
oc patch OperatorHub cluster --type json \
  -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": true}]'
```

### 2. Create Catalog Source

```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: my-operator-manifests
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: quay.io/${USERNAME}/rhpam-operator-index:8.0.9
  displayName: My Operator Catalog
  publisher: grpc
EOF
```

### 3. Create Target Namespace

```bash
oc create namespace test-deploy
```

### 4. Create OperatorGroup

```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: test-deploy-og
  namespace: test-deploy
spec:
  targetNamespaces:
  - test-deploy
EOF
```

**Note:** OperatorGroup is required for OLM to install operators in a namespace.

### 5. Create Subscription

```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: bamoe-businessautomation-operator
  namespace: test-deploy
spec:
  channel: 8.x-stable
  name: bamoe-businessautomation-operator
  source: my-operator-manifests
  sourceNamespace: openshift-marketplace
EOF
```

**Important:** Use channel `8.x-stable`, not `stable`

---

## Verification Steps

### 1. Check Catalog Source

```bash
# Verify catalog source is running
oc get catalogsource -n openshift-marketplace
oc get pods -n openshift-marketplace | grep my-operator

# Expected: Pod in Running state
```

### 2. Check Package Availability

```bash
oc get packagemanifest bamoe-businessautomation-operator -n test-deploy
```

### 3. Check Subscription Status

```bash
oc get subscription bamoe-businessautomation-operator -n test-deploy
oc describe subscription bamoe-businessautomation-operator -n test-deploy
```

### 4. Check Install Plan

```bash
oc get installplan -n test-deploy
```

### 5. Check CSV (ClusterServiceVersion)

```bash
oc get csv -n test-deploy

# Expected output:
# NAME                                                       DISPLAY                         VERSION              PHASE
# bamoe-businessautomation-operator.8.0.9-1-dev-22hf4cvz47   IBM Business Automation (DEV)   8.0.9-1+22hf4cvz47   Succeeded
```

### 6. Check Operator Pod

```bash
oc get pods -n test-deploy

# Expected: bamoe-business-automation-operator pod in Running state
```

### 7. Verify KieApp API

```bash
oc api-resources --api-group=app.kiegroup.org

# Expected output:
# NAME      SHORTNAMES   APIVERSION            NAMESPACED   KIND
# kieapps                app.kiegroup.org/v2   true         KieApp
```

---

## Issues and Resolutions

### Issue 1: Missing KieApp API

**Symptom:**
- Operator installed but KieApp CRD not available
- Cannot create KieApp resources

**Root Cause:**
- Old bundle/index images were cached locally
- Bundle didn't include the CRD properly

**Resolution:**
```bash
# 1. Remove old images
podman rmi -f $(podman images -q)

# 2. Rebuild bundle and index (see Build Process section)

# 3. Recreate catalog source
oc delete catalogsource my-operator-manifests -n openshift-marketplace
# Then create new catalog source with updated index
```

### Issue 2: Operator Deployment Timeout

**Symptom:**
```
install failed: deployment bamoe-business-automation-operator not ready before timeout
```

**Root Cause:**
- Operator image not available (ImagePullBackOff)
- CSV referenced `quay.io/kiegroup/kie-cloud-operator:8.0.9` which doesn't exist

**Resolution:**
1. Build operator image locally (see Build Process)
2. Push to your registry
3. Update CSV to reference your image
4. Rebuild and push bundle/index

### Issue 3: Wrong Subscription Channel

**Symptom:**
- Subscription created but no install plan generated

**Root Cause:**
- Used channel `stable` instead of `8.x-stable`

**Resolution:**
```bash
# Delete and recreate subscription with correct channel
oc delete subscription bamoe-businessautomation-operator -n test-deploy

cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: bamoe-businessautomation-operator
  namespace: test-deploy
spec:
  channel: 8.x-stable  # Correct channel
  name: bamoe-businessautomation-operator
  source: my-operator-manifests
  sourceNamespace: openshift-marketplace
EOF
```

### Issue 4: Missing OperatorGroup

**Symptom:**
- Subscription exists but no install plan created
- No operator deployment

**Root Cause:**
- OperatorGroup is required for OLM to install operators

**Resolution:**
```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: test-deploy-og
  namespace: test-deploy
spec:
  targetNamespaces:
  - test-deploy
EOF
```

---

## KieApp Deployment

### 1. Create Pull Secret for Artifactory

**Required for pulling BAMOE images from IBM Artifactory**

```bash
# Create pull secret
oc create secret docker-registry docker-artifactory \
  --docker-server=na.artifactory.swg-devops.com \
  --docker-username=<your-email> \
  --docker-password=<your-token> \
  -n test-deploy

# Link to default service account
oc secrets link default docker-artifactory --for=pull -n test-deploy
```

**Important:** The docker-server should be `na.artifactory.swg-devops.com` (without `https://`)

### 2. Create KieApp with Reduced Resources (for CRC)

```bash
cat <<EOF | oc apply -f -
apiVersion: app.kiegroup.org/v2
kind: KieApp
metadata:
  name: rhpam-trial
  namespace: test-deploy
spec:
  environment: rhpam-trial
  objects:
    console:
      resources:
        limits:
          memory: "2Gi"
        requests:
          memory: "1Gi"
    servers:
    - resources:
        limits:
          memory: "1Gi"
        requests:
          memory: "512Mi"
EOF
```

### 3. Monitor KieApp Deployment

```bash
# Check KieApp status
oc get kieapp rhpam-trial -n test-deploy

# Check pods
oc get pods -n test-deploy

# Check pod details if issues
oc describe pod <pod-name> -n test-deploy
```

### Issue 5: KieApp Stuck in Provisioning - Insufficient Memory

**Symptom:**
```
0/1 nodes are available: 1 Insufficient memory
```

**Resolution:**
- Use reduced resource requests (shown in example above)
- Or increase CRC memory:
```bash
crc stop
crc config set memory 16384  # 16GB
crc start
```

### Issue 6: KieApp Stuck in Provisioning - Image Pull Authentication

**Symptom:**
```
Failed to pull image: unable to retrieve auth token: invalid username/password
```

**Root Cause:**
- Operator creates service account `rhpam-trial-rhpamsvc`
- Service account doesn't have the pull secret
- Manually added secrets don't persist (operator manages the SA)

**Current Status:**
- Pull secret exists and credentials are valid
- Service account doesn't retain manually added imagePullSecrets
- Operator doesn't support specifying imagePullSecrets in KieApp CR

**Recommended Solutions:**

#### Option 1: Global Pull Secret (Recommended for CRC)

```bash
# Get current global pull secret
oc get secret/pull-secret -n openshift-config \
  -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d > /tmp/global-pull-secret.json

# Add artifactory credentials
oc registry login \
  --registry="na.artifactory.swg-devops.com" \
  --auth-basic="<username>:<password>" \
  --to=/tmp/global-pull-secret.json

# Update global pull secret
oc set data secret/pull-secret -n openshift-config \
  --from-file=.dockerconfigjson=/tmp/global-pull-secret.json

# Wait for nodes to update (may take a few minutes)
# Then recreate KieApp
oc delete kieapp rhpam-trial -n test-deploy
oc create -f <kieapp-yaml>
```

#### Option 2: Modify Operator (Advanced)

Add support for `imagePullSecrets` in KieApp CR spec, requiring operator code changes and rebuild.

#### Option 3: MutatingWebhookConfiguration (Advanced)

Create a webhook to automatically inject pull secrets into pods.

---

## Cleanup

### Remove KieApp

```bash
oc delete kieapp rhpam-trial -n test-deploy
```

### Remove Operator

```bash
# Delete subscription
oc delete subscription bamoe-businessautomation-operator -n test-deploy

# Delete CSV
oc delete csv -n test-deploy --all

# Delete catalog source
oc delete catalogsource my-operator-manifests -n openshift-marketplace
```

### Restore Default Catalog Sources

```bash
oc patch OperatorHub cluster --type json \
  -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": false}]'
```

---

## Quick Reference Commands

### Check Operator Status
```bash
oc get csv -n test-deploy
oc get pods -n test-deploy
oc logs -n test-deploy deployment/bamoe-business-automation-operator
```

### Check KieApp Status
```bash
oc get kieapp -n test-deploy
oc describe kieapp rhpam-trial -n test-deploy
```

### Debug Image Pull Issues
```bash
# Check service account
oc get sa rhpam-trial-rhpamsvc -n test-deploy -o yaml

# Check pod image pull secrets
oc get pod <pod-name> -n test-deploy -o jsonpath='{.spec.imagePullSecrets}'

# Check pod events
oc describe pod <pod-name> -n test-deploy | grep -A 20 Events:
```

### Test Registry Credentials
```bash
podman login na.artifactory.swg-devops.com \
  --username <username> \
  --password <password>
```

---

## Summary

This guide covers the complete process of building, deploying, and troubleshooting the KIE Cloud Operator on OpenShift/CRC. The main challenges encountered were:

1. **Image availability** - Required building and pushing custom operator image
2. **Bundle/Index updates** - Needed to rebuild after CSV changes
3. **OLM requirements** - OperatorGroup and correct channel name
4. **Image pull authentication** - Ongoing issue requiring global pull secret configuration

For production deployments, ensure:
- All images are available in accessible registries
- Pull secrets are properly configured
- Sufficient cluster resources are available
- Operator is deployed via official channels when available