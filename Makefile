.PHONY: all build test test-integration test-smoke smoke lint fmt install ci clean docker

BINARY_NAME=ninjops
MAIN_PATH=./cmd/ninjops
GO=go
GOFLAGS=-v
GOBIN_DIR=$(shell $(GO) env GOPATH)/bin

all: build

build:
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)

test:
	$(GO) test $(GOFLAGS) -race -cover ./...

test-integration:
	$(GO) test $(GOFLAGS) -race -cover -tags=integration ./...

test-smoke:
	$(GO) test $(GOFLAGS) ./internal/app -run TestCLISmoke -count=1

smoke: test-smoke

test-coverage:
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout 5m

fmt:
	$(GO) fmt ./...
	gofmt -s -w .

install: build
	mkdir -p $(GOBIN_DIR)
	cp bin/$(BINARY_NAME) $(GOBIN_DIR)/

ci: fmt lint test

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

docker:
	docker build -t ninjops:latest .

docker-run:
	docker run -p 8080:8080 ninjops:latest

.PHONY: check
check: fmt lint test
