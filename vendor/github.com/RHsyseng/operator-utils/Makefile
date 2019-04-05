.PHONY: all
all: test

.PHONY: dep
dep:
	dep ensure -v

.PHONY: format
format: dep
	go fmt ./...

.PHONY: vet
vet: format
	go vet ./...

.PHONY: test
test: vet
	go test ./...
