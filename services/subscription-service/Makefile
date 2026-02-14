# ==============================================================================
# StreamVault Subscription Service - Makefile
# ==============================================================================

.PHONY: build run test docker-build docker-run clean restore migrate

# Variables
PROJECT_NAME = SubscriptionService
API_PROJECT = src/SubscriptionService.Api/SubscriptionService.Api.csproj
DOCKER_IMAGE = streamvault/subscription-service
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
	rm -rf src/*/bin src/*/obj

# ---- Docker ----

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run -d \
		--name subscription-service \
		-p 5007:5007 \
		-e ConnectionStrings__DefaultConnection="Host=host.docker.internal;Port=5432;Database=streamvault_subscriptions;Username=streamvault;Password=streamvault" \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	docker stop subscription-service && docker rm subscription-service

# ---- Database Migrations ----

migrate:
	dotnet ef migrations add $(name) \
		--project src/SubscriptionService.Infrastructure \
		--startup-project src/SubscriptionService.Api \
		--output-dir Persistence/Migrations

migrate-apply:
	dotnet ef database update \
		--project src/SubscriptionService.Infrastructure \
		--startup-project src/SubscriptionService.Api

# ---- Utilities ----

format:
	dotnet format $(API_PROJECT)

lint:
	dotnet format $(API_PROJECT) --verify-no-changes
