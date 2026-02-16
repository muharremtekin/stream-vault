# ==============================================================================
# StreamVault User Service - Multi-stage .NET 8 Dockerfile
# ==============================================================================

# Stage 1: Build
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project files and restore dependencies
COPY src/UserService.Domain/*.csproj UserService.Domain/
COPY src/UserService.Application/*.csproj UserService.Application/
COPY src/UserService.Infrastructure/*.csproj UserService.Infrastructure/
COPY src/UserService.Api/*.csproj UserService.Api/
RUN dotnet restore UserService.Api/UserService.Api.csproj

# Copy all source code and build
COPY src/ .
RUN dotnet build UserService.Api/UserService.Api.csproj -c Release --no-restore

# Stage 2: Publish
FROM build AS publish
RUN dotnet publish UserService.Api/UserService.Api.csproj -c Release -o /app/publish --no-build

# Stage 3: Runtime
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends wget && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser

COPY --from=publish /app/publish .

# Set environment variables
ENV ASPNETCORE_URLS=http://+:8080
ENV ASPNETCORE_ENVIRONMENT=Production
ENV DOTNET_RUNNING_IN_CONTAINER=true

EXPOSE 8080

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --spider --quiet http://localhost:8080/health/live || exit 1

ENTRYPOINT ["dotnet", "UserService.Api.dll"]
