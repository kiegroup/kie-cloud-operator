#!/bin/sh

source ./hack/go-mod-env.sh

REPO_LINK=$(echo ${JOB_SPEC} | python -c "import sys, json; print(json.load(sys.stdin)['refs']['repo_link'])")
if [[ ${REPO_LINK} != "https://github.com/kiegroup/kie-cloud-operator" ]]; then
    echo "repo_link not correct, wrong repo for test"
    exit 0
fi
BASE_SHA=$(echo ${JOB_SPEC} | python -c "import sys, json; print(json.load(sys.stdin)['refs']['base_sha'])")
if [[ -z ${BASE_SHA} ]]; then
    echo "base_sha not found, can't execute diff"
    exit 1
fi
git remote add origin ${REPO_LINK}
git fetch origin ${BASE_SHA}
VERSION=$(go run getversion.go)
RESULT=$(git diff --name-only ${BASE_SHA} | grep "^rhpam-config/" | grep -v "^rhpam-config/${VERSION}")
if [[ ${RESULT} ]]; then
    echo "Detected changes to an older version's config file(s). Current version changes are only allowed in rhpam-config/${VERSION}."
    echo "Undo changes to the following files -"
    echo "${RESULT}"
    exit 1
fi
