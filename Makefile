.PHONY: clean checks test build image dependencies

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

LEGO_IMAGE := xenolf/lego

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
VERSION := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))

default: clean checks test build

clean:
	rm -rf dist/ builds/ cover.out

build: clean
	@echo Version: $(VERSION)
	go build -v -ldflags '-X "main.version=${VERSION}"'

dependencies:
	dep ensure -v

test: clean
	go test -v -cover ./...

e2e: clean
	LEGO_E2E_TESTS=local go test -count=1 -v ./e2e/...

checks:
	golangci-lint run

fmt:
	gofmt -s -l -w $(SRCS)
image:
	@echo Version: $(VERSION)
	docker build -t $(LEGO_IMAGE) .
