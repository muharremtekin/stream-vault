.PHONY: build test lint proto clean run docker-build help

BINARY_NAME = encoding-service

build:
	cargo build --release

test:
	cargo test --all-features

lint:
	cargo clippy -- -D warnings
	cargo fmt --check

proto:
	cd ../.. && bash scripts/generate-proto.sh

clean:
	cargo clean

run: build
	./target/release/$(BINARY_NAME)

docker-build:
	cd ../.. && docker build -f services/encoding-service/Dockerfile -t streamvault/encoding-service:latest services/encoding-service/

help:
	@echo "Available targets:"
	@echo "  build        - Build release binary"
	@echo "  test         - Run tests"
	@echo "  lint         - Run clippy and fmt check"
	@echo "  proto        - Generate proto code (via root script)"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run"
	@echo "  docker-build - Build Docker image"
