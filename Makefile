.PHONY: build test lint install clean check

build:
	go build -o bin/deadfunc ./cmd/deadfunc

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

install:
	go install ./cmd/deadfunc

clean:
	rm -rf bin/ dist/

check: test build lint
