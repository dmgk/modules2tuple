all: build

build:
	go build

test:
	go test -tags=online -count=1 ./...

install:
	go install

clean:
	go clean

.PHONY: all build test install
