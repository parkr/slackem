all: build test

deps:
	go get github.com/stretchr/testify/assert \
	  github.com/nlopes/slack

fmt:
	go fmt

build: deps fmt
	go build

test: deps
	go test -v -cover
