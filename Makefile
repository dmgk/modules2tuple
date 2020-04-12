all: build

build:
	@go build

test: all
	go test -tags=online ./...
	./e2e.sh

install:
	@go install

clean:
	@go clean

.PHONY: all build test install
