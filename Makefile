.PHONY: all

GOFILES := $(shell go list -f '{{range $$index, $$element := .GoFiles}}{{$$.Dir}}/{{$$element}}{{"\n"}}{{end}}' ./... | grep -v '/vendor/')

default: clean checks test build

clean:
	rm -rf dist/ builds/ cover.out

build: clean
	go build


dependencies:
	dep ensure -v

test: clean
	go test -v -cover ./...

checks: check-fmt
	gometalinter ./...

check-fmt: SHELL := /bin/bash
check-fmt:
	diff -u <(echo -n) <(gofmt -d $(GOFILES))
