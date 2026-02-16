# ==============================================================================
# StreamVault Subscription Service - Multi-stage .NET 8 Dockerfile
# ==============================================================================

# Stage 1: Build
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project files and restore dependencies
COPY src/SubscriptionService.Domain/*.csproj SubscriptionService.Domain/
COPY src/SubscriptionService.Application/*.csproj SubscriptionService.Application/
COPY src/SubscriptionService.Infrastructure/*.csproj SubscriptionService.Infrastructure/
COPY src/SubscriptionService.Api/*.csproj SubscriptionService.Api/
RUN dotnet restore SubscriptionService.Api/SubscriptionService.Api.csproj

# Copy all source code and build
COPY src/ .
RUN dotnet build SubscriptionService.Api/SubscriptionService.Api.csproj -c Release --no-restore

# Stage 2: Publish
FROM build AS publish
RUN dotnet publish SubscriptionService.Api/SubscriptionService.Api.csproj -c Release -o /app/publish --no-build

# Stage 3: Runtime
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends wget && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser

COPY --from=publish /app/publish .

# Set environment variables
ENV ASPNETCORE_URLS=http://+:5007
ENV ASPNETCORE_ENVIRONMENT=Production
ENV DOTNET_RUNNING_IN_CONTAINER=true

EXPOSE 5007

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --spider --quiet http://localhost:5007/health/live || exit 1

ENTRYPOINT ["dotnet", "SubscriptionService.Api.dll"]
