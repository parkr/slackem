all: build test

fmt:
	go fmt

build: fmt
	go build

test:
	go test -v -cover
