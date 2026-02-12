# ==============================================================================
# StreamVault User Service - Makefile
# ==============================================================================

.PHONY: build run test docker-build docker-run clean restore migrate

# Variables
PROJECT_NAME = UserService
API_PROJECT = src/UserService.Api/UserService.Api.csproj
UNIT_TESTS = tests/UserService.UnitTests/UserService.UnitTests.csproj
INTEGRATION_TESTS = tests/UserService.IntegrationTests/UserService.IntegrationTests.csproj
DOCKER_IMAGE = streamvault/user-service
DOCKER_TAG = latest
CONFIGURATION = Release

# ---- Build & Run ----

build:
	dotnet build $(API_PROJECT) -c $(CONFIGURATION)

run:
	dotnet run --project $(API_PROJECT) --launch-profile http

run-watch:
	dotnet watch run --project $(API_PROJECT)

restore:
	dotnet restore $(API_PROJECT)

clean:
	dotnet clean $(API_PROJECT)
	rm -rf src/*/bin src/*/obj tests/*/bin tests/*/obj

# ---- Testing ----

test: test-unit test-integration

test-unit:
	dotnet test $(UNIT_TESTS) -c $(CONFIGURATION) --verbosity normal

test-integration:
	dotnet test $(INTEGRATION_TESTS) -c $(CONFIGURATION) --verbosity normal

test-coverage:
	dotnet test $(UNIT_TESTS) -c $(CONFIGURATION) \
		--collect:"XPlat Code Coverage" \
		--results-directory ./coverage

# ---- Docker ----

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run -d \
		--name user-service \
		-p 5100:8080 \
		-e ConnectionStrings__DefaultConnection="Host=host.docker.internal;Port=5432;Database=streamvault_users;Username=postgres;Password=postgres" \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	docker stop user-service && docker rm user-service

# ---- Database Migrations ----

migrate:
	dotnet ef migrations add $(name) \
		--project src/UserService.Infrastructure \
		--startup-project src/UserService.Api \
		--output-dir Migrations

migrate-apply:
	dotnet ef database update \
		--project src/UserService.Infrastructure \
		--startup-project src/UserService.Api

# ---- Utilities ----

format:
	dotnet format $(API_PROJECT)

lint:
	dotnet format $(API_PROJECT) --verify-no-changes
