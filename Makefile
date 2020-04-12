all: build

build:
	@go build

test:
	go test -tags=online,e2e ./...

install:
	@go install

clean:
	@go clean

.PHONY: all build test install
