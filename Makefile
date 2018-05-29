.PHONY: all

default: clean checks test build

test: clean
	go test -v -cover ./...

clean:
	rm -rf dist/ builds/ cover.out

checks:
	go vet ./...

build: clean
	go build
