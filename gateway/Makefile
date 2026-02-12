APP_NAME    := gateway
MODULE      := github.com/streamvault/gateway
BUILD_DIR   := ./build
BINARY      := $(BUILD_DIR)/$(APP_NAME)
DOCKER_TAG  := streamvault/$(APP_NAME):latest
GO          := go
GOFLAGS     := -v
LDFLAGS     := -s -w

.PHONY: build run test docker-build lint clean

## build: Compile the gateway binary into ./build/
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/gateway

## run: Build and run the gateway locally
run: build
	$(BINARY) -config config.yaml

## test: Run all unit tests with race detection and coverage
test:
	$(GO) test $(GOFLAGS) -race -cover ./...

## docker-build: Build the Docker image using the multi-stage Dockerfile
docker-build:
	docker build -t $(DOCKER_TAG) .

## lint: Run golangci-lint (must be installed separately)
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## help: Print this help message
help:
	@echo "Available targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
