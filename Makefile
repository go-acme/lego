.PHONY: all

default: clean checks test

test: clean
	go test -v -cover ./...

clean:
	rm -rf dist/ builds/ cover.out

checks:
	go vet ./...

builds: clean
	./releaser.sh
