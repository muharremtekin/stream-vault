.PHONY: build test lint clean run

BINARY=notification
CMD_DIR=./cmd/notification

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(BINARY) $(CMD_DIR)

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

run: build
	./bin/$(BINARY)
