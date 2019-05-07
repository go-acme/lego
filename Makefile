.PHONY: clean checks test build image dependencies e2e fmt

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

LEGO_IMAGE := go-acme/lego
MAIN_DIRECTORY := ./cmd/lego/
ifeq (${GOOS}, windows)
    BIN_OUTPUT := dist/lego.exe
else
    BIN_OUTPUT := dist/lego
endif

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse HEAD)
VERSION := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))

default: clean generate-dns checks test build

clean:
	rm -rf dist/ builds/ cover.out

build: clean
	@echo Version: $(VERSION)
	go build -v -ldflags '-X "main.version=${VERSION}"' -o ${BIN_OUTPUT} ${MAIN_DIRECTORY}

image:
	@echo Version: $(VERSION)
	docker build -t $(LEGO_IMAGE) .

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

# Release helper
.PHONY: patch minor major detach

patch:
	go run internal/release.go release -m patch

minor:
	go run internal/release.go release -m minor

major:
	go run internal/release.go release -m major

detach:
	go run internal/release.go detach

# Docs
.PHONY: docs-build docs-serve docs-themes

docs-build: generate-dns
	@make -C ./docs hugo-build

docs-serve: generate-dns
	@make -C ./docs hugo

docs-themes:
	@make -C ./docs hugo-themes

# DNS Documentation
.PHONY: generate-dns validate-doc

generate-dns:
	go generate ./...

validate-doc: generate-dns
ifneq ($(shell git status --porcelain -- ./docs/ ./cmd/ 2>/dev/null),)
	@echo 'The documentation must be regenerated, please use `make generate-dns`.'
	@git status --porcelain -- ./docs/ ./cmd/ 2>/dev/null
	@exit 2
else
	@echo 'All documentation changes are done the right way.'
endif
