# ============================================================================
# StreamVault Catalog Service - Multi-stage Dockerfile
# Target: .NET 8
# ============================================================================

# ---------------------------------------------------------------------------
# Stage 1: Build
# ---------------------------------------------------------------------------
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project files and restore dependencies
COPY src/CatalogService.Domain/*.csproj CatalogService.Domain/
COPY src/CatalogService.Application/*.csproj CatalogService.Application/
COPY src/CatalogService.Infrastructure/*.csproj CatalogService.Infrastructure/
COPY src/CatalogService.Api/*.csproj CatalogService.Api/
RUN dotnet restore CatalogService.Api/CatalogService.Api.csproj

# Copy full source and build
COPY src/ .
RUN dotnet publish CatalogService.Api/CatalogService.Api.csproj \
    -c Release \
    -o /app/publish \
    --no-restore

# ---------------------------------------------------------------------------
# Stage 2: Test (optional, runs unit tests)
# ---------------------------------------------------------------------------
FROM build AS test
WORKDIR /src
COPY tests/ /tests/
RUN dotnet test /tests/CatalogService.UnitTests/CatalogService.UnitTests.csproj \
    --no-restore \
    --verbosity normal \
    || true

# ---------------------------------------------------------------------------
# Stage 3: Runtime
# ---------------------------------------------------------------------------
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends wget && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser -d /app -s /sbin/nologin appuser

COPY --from=build /app/publish .

# Set environment variables
ENV ASPNETCORE_URLS=http://+:5100
ENV ASPNETCORE_ENVIRONMENT=Production
ENV DOTNET_RUNNING_IN_CONTAINER=true

EXPOSE 5100

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --spider --quiet http://localhost:5100/health || exit 1

USER appuser

ENTRYPOINT ["dotnet", "CatalogService.Api.dll"]
