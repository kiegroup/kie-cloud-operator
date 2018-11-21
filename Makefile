# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

IMAGE = quay.io/kiegroup/kie-cloud-operator
#VERSION = $(shell git describe --dirty --tags --always)
#REPO = github.com/kiegroup/kie-cloud-operator
#BUILD_PATH = $(REPO)/cmd/manager

#export CGO_ENABLED:=0

all: build

dep:
	$(Q)dep ensure -v

format:
	$(Q)go fmt ./...

go-generate: dep
	$(Q)go generate ./...

sdk-generate: dep
	operator-sdk generate k8s

vet: sdk-generate
	$(Q)go vet ./...

test: vet format
	$(Q)go test ./...

build: go-generate test
	operator-sdk build $(IMAGE)

clean:
	rm -rf build/_output pkg/controller/kieapp/defaults/a_defaults-packr.go

.PHONY: all dep vet go-generate sdk-generate format test build clean

# test/ci-go: test/sanity test/unit test/subcommand test/e2e/go

# test/ci-ansible: test/e2e/ansible

# test/sanity:
# 	./hack/tests/sanity-check.sh

# test/unit:
# 	./hack/tests/unit.sh

# test/subcommand:
# 	./hack/tests/test-subcommand.sh

# test/e2e: test/e2e/go test/e2e/ansible

# test/e2e/go:
# 	./hack/tests/e2e-go.sh

# test/e2e/ansible:
# 	./hack/tests/e2e-ansible.sh
