# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

#VERSION = $(shell git describe --dirty --tags --always)
#REPO = github.com/kiegroup/kie-cloud-operator
#BUILD_PATH = $(REPO)/cmd/manager

#export CGO_ENABLED:=0

.PHONY: all
all: build

.PHONY: dep
dep:
	./hack/go-dep.sh

.PHONY: format
format:
	./hack/go-fmt.sh

.PHONY: go-generate
go-generate: dep
	$(Q)go generate ./...

.PHONY: sdk-generate
sdk-generate: dep
	operator-sdk generate k8s

.PHONY: vet
vet:
	./hack/go-vet.sh

.PHONY: test
test:
	./hack/go-test.sh

.PHONY: lint
lint:
	# Temporarily disabled
	# ./hack/go-lint.sh
	# ./hack/yaml-lint.sh

.PHONY: build
build:
	./hack/go-build.sh

.PHONY: clean
clean:
	rm -rf build/_output pkg/controller/kieapp/defaults/a_defaults-packr.go

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
