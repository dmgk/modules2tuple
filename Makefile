all: build

build:
	@go build

test:
	@if [ -z "$$M2T_GITHUB" ]; then \
		echo "*** Please set M2T_GITHUB=<github_username>:<personal_access_token>"; \
		exit 1; \
	fi
	go test -tags=online,e2e ./...

install:
	@go install

clean:
	@go clean

.PHONY: all build test install
