all: build

build:
	@go build

test: all
	go test -tags=online ./...
	./test.sh

install:
	@go install

clean:
	@go clean

.PHONY: all build test install
