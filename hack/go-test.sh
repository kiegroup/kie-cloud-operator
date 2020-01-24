#!/bin/sh

set -x
grep --version
git version

. ./hack/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./hack/go-vet.sh
    ./hack/go-fmt.sh
else
    #JOB_SPEC='{"type":"presubmit","job":"pull-ci-kiegroup-kie-cloud-operator-master-unit","buildid":"659","prowjobid":"c581cca6-3ee0-11ea-bc44-0a58ac103085","refs":{"org":"kiegroup","repo":"kie-cloud-operator","repo_link":"https://github.com/kiegroup/kie-cloud-operator","base_ref":"master","base_sha":"6c7e0915536125ba931b9df30c349c008f142a8b","base_link":"https://github.com/kiegroup/kie-cloud-operator/commit/6c7e0915536125ba931b9df30c349c008f142a8b","pulls":[{"number":344,"author":"tchughesiv","sha":"253a9f2c35747295c0a35df3afa5b599de9f1e6f","link":"https://github.com/kiegroup/kie-cloud-operator/pull/344","commit_link":"https://github.com/kiegroup/kie-cloud-operator/pull/344/commits/253a9f2c35747295c0a35df3afa5b599de9f1e6f","author_link":"https://github.com/tchughesiv"}]}}'
    BASE_REF=$(echo ${JOB_SPEC} | python -c "import sys, json; print(json.load(sys.stdin)['refs']['base_ref'])")
    if [[ -z ${BASE_REF} ]]; then
        echo "base_ref not found, can't execute diff"
        exit 1
    else
        VERSION=$(go run getversion.go -product)
        RESULT=$(git diff --name-only origin/${BASE_REF} | grep "^config/" | grep -v "^config/${VERSION}")
        if [[ ${RESULT} ]]; then
            echo "\nDetected changes to an older version's config file(s). Current version changes are only allowed in config/${VERSION}."
            echo "Undo changes to the following files -"
            echo "${RESULT}\n"
            exit 1
        fi
    fi
fi

setGoModEnv

go test -mod=vendor -count=1 ./...

exit 1