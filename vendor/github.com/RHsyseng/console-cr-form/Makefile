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
#REPO = github.com/RHsyseng/console-cr-form

#export CGO_ENABLED:=0

.PHONY: all
all: build

.PHONY: npm
npm:
	rm -f frontend/package-lock.json
	cd frontend; npm install
	npm --prefix frontend run build

.PHONY: dep
dep:
	dep ensure -v

.PHONY: go-generate
go-generate: dep
	$(Q)go generate ./...

.PHONY: build
build: npm go-generate
	CGO_ENABLED=0 go build -v -a -o build/console-cr-form github.com/RHsyseng/console-cr-form/cmd

.PHONY: clean
clean:
	rm -rf build \
		pkg/web/packrd \
		pkg/web/web-packr.go
