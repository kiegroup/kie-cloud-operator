# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

# Container runtime engine used for building the images
# set BUILDER env to define the builder, defaults to podman, can be podman, buildah or docker
BUILDER ?= podman

.PHONY: all
all: build

.PHONY: mod
mod:
	./hack/go-mod.sh

.PHONY: format
format:
	./hack/go-fmt.sh

.PHONY: go-generate
go-generate: mod
	./hack/go-gen.sh

.PHONY: sdk-generate
sdk-generate: mod
	./hack/go-sdk-gen.sh

.PHONY: vet
vet:
	./hack/go-vet.sh

.PHONY: test
test:
	./hack/go-test.sh

.PHONY: check-configs
check-configs:
	./hack/go-configs.sh

.PHONY: lint
lint:
	# Temporarily disabled
	# ./hack/go-lint.sh
	# ./hack/yaml-lint.sh

.PHONY: build
build:
	./hack/go-build.sh ${BUILDER}

.PHONY: bundle
bundle:
	LOCAL=true ./hack/go-build-bundle.sh ${BUILDER}

.PHONY: bundle-scratch
bundle-scratch:
	./hack/go-build-bundle.sh ${BUILDER}

.PHONY: bundle-release
bundle-release:
	./hack/go-build-bundle.sh ${BUILDER} release

.PHONY: bundle-dev
bundle-dev:
	DEV=true LOCAL=true ./hack/go-build-bundle.sh ${BUILDER}

.PHONY: rhel
rhel:
	LOCAL=true ./hack/go-build.sh ${BUILDER} rhel

.PHONY: rhel-scratch
rhel-scratch:
	./hack/go-build.sh rhel

.PHONY: rhel-release
rhel-release:
	./hack/go-build.sh ${BUILDER} rhel release

.PHONY: rhel-nightly
rhel-nightly:
	./hack/go-build.sh ${BUILDER} rhel nightly

.PHONY: csv
csv: sdk-generate
	./hack/go-csv.sh

.PHONY: clean
clean:
	rm -rf build/_output \
		pkg/controller/kieapp/defaults/defaults-packr.go \
		pkg/controller/kieapp/defaults/packrd \
		pkg/controller/kieapp/kieapp-packr.go \
		pkg/controller/kieapp/packrd \
		pkg/ui/ui-packr.go \
		pkg/ui/packrd \
		modules/builder/kie-cloud-operator.tar.gz \
		target/

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
