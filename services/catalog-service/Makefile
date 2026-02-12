# ============================================================================
# StreamVault Catalog Service - Makefile
# ============================================================================

.PHONY: build run test clean restore docker-build docker-run publish lint help

SERVICE_NAME   := catalog-service
DOCKER_IMAGE   := streamvault/$(SERVICE_NAME)
DOCKER_TAG     := latest
API_PROJECT    := src/CatalogService.Api/CatalogService.Api.csproj
TEST_PROJECT   := tests/CatalogService.UnitTests/CatalogService.UnitTests.csproj
CONFIGURATION  := Release

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

restore: ## Restore NuGet packages
	dotnet restore $(API_PROJECT)

build: restore ## Build the solution
	dotnet build $(API_PROJECT) -c $(CONFIGURATION) --no-restore

run: ## Run the API locally
	dotnet run --project $(API_PROJECT) --launch-profile "http"

watch: ## Run with hot reload
	dotnet watch run --project $(API_PROJECT)

test: ## Run unit tests
	dotnet test $(TEST_PROJECT) --verbosity normal

test-coverage: ## Run tests with coverage report
	dotnet test $(TEST_PROJECT) \
		--collect:"XPlat Code Coverage" \
		--results-directory ./coverage

publish: ## Publish release build
	dotnet publish $(API_PROJECT) -c $(CONFIGURATION) -o ./publish

clean: ## Clean build artifacts
	dotnet clean $(API_PROJECT)
	rm -rf ./publish ./coverage

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -d \
		--name $(SERVICE_NAME) \
		-p 5100:5100 \
		-e MongoDB__ConnectionString=mongodb://host.docker.internal:27017 \
		-e Consul__Address=http://host.docker.internal:8500 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## Stop Docker container
	docker stop $(SERVICE_NAME) && docker rm $(SERVICE_NAME)

lint: ## Run code analysis
	dotnet format $(API_PROJECT) --verify-no-changes

format: ## Format code
	dotnet format $(API_PROJECT)
