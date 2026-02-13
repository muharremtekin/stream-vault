.PHONY: build test lint proto clean run

BINARY=streaming
CMD_DIR=./cmd/streaming

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY) $(CMD_DIR)

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

proto:
	cd ../.. && bash scripts/generate-proto.sh

clean:
	rm -rf bin/

run: build
	./bin/$(BINARY)
