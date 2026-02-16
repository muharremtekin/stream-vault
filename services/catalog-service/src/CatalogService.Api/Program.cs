using CatalogService.Api.Middleware;
using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Application.Mappings;
using CatalogService.Infrastructure;
using CatalogService.Infrastructure.Seed;
using Consul;
using FluentValidation;
using FluentValidation.AspNetCore;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Prometheus;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using Serilog;
using Serilog.Formatting.Compact;

var builder = WebApplication.CreateBuilder(args);

// Graceful shutdown timeout
builder.Services.Configure<HostOptions>(opts =>
{
    opts.ShutdownTimeout = TimeSpan.FromSeconds(30);
});

// Serilog
builder.Host.UseSerilog((context, loggerConfig) =>
{
    loggerConfig
        .ReadFrom.Configuration(context.Configuration)
        .Enrich.FromLogContext()
        .Enrich.WithProperty("ServiceName", "catalog-service")
        .WriteTo.Console(new RenderedCompactJsonFormatter());
});

// OpenTelemetry
var otelEndpoint = builder.Configuration["OpenTelemetry:Endpoint"];
if (!string.IsNullOrEmpty(otelEndpoint))
{
    builder.Services.AddOpenTelemetry()
        .WithTracing(tracing =>
        {
            tracing
                .SetResourceBuilder(ResourceBuilder.CreateDefault()
                    .AddService("catalog-service", serviceVersion: "1.0.0"))
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddSource("MongoDB.Driver")
                .AddOtlpExporter(opts =>
                {
                    opts.Endpoint = new Uri(otelEndpoint);
                });
        });
}

// ---------------------------------------------------------------------------
// Services
// ---------------------------------------------------------------------------

// Controllers
builder.Services.AddControllers();

// Swagger / OpenAPI
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(options =>
{
    options.SwaggerDoc("v1", new Microsoft.OpenApi.Models.OpenApiInfo
    {
        Title = "StreamVault Catalog Service",
        Version = "v1",
        Description = "Catalog microservice for StreamVault streaming platform."
    });
});

// MediatR
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssemblyContaining<CreateMovieCommand>());

// AutoMapper
builder.Services.AddAutoMapper(typeof(MappingProfile).Assembly);

// FluentValidation
builder.Services.AddFluentValidationAutoValidation();
builder.Services.AddValidatorsFromAssemblyContaining<CreateMovieValidator>();

// Infrastructure (MongoDB, Repositories)
builder.Services.AddInfrastructure(builder.Configuration);

// Consul service discovery
builder.Services.AddSingleton<IConsulClient, ConsulClient>(sp =>
    new ConsulClient(config =>
    {
        var consulAddress = builder.Configuration.GetValue<string>("Consul:Address") ?? "http://localhost:8500";
        config.Address = new Uri(consulAddress);
    }));

// Health checks
builder.Services.AddHealthChecks()
    .AddMongoDb(
        builder.Configuration.GetValue<string>("MongoDB:ConnectionString") ?? "mongodb://localhost:27017",
        name: "mongodb",
        tags: new[] { "ready" });

// CORS
builder.Services.AddCors(options =>
{
    options.AddPolicy("AllowAll", policy =>
    {
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

var app = builder.Build();

// ---------------------------------------------------------------------------
// Pipeline
// ---------------------------------------------------------------------------

app.UseCorrelationId();
app.UseSerilogRequestLogging();
app.UseExceptionHandling();
app.UseHttpMetrics();

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI(options =>
    {
        options.SwaggerEndpoint("/swagger/v1/swagger.json", "Catalog Service v1");
    });
}

app.UseCors("AllowAll");

app.UseAuthorization();

app.MapControllers();
app.MapHealthChecks("/health/live", new HealthCheckOptions { Predicate = _ => false });
app.MapHealthChecks("/health/ready", new HealthCheckOptions
{
    Predicate = check => check.Tags.Contains("ready")
});
app.MapHealthChecks("/health/startup", new HealthCheckOptions
{
    Predicate = check => check.Tags.Contains("ready")
});
app.MapHealthChecks("/health", new HealthCheckOptions { Predicate = _ => false });
app.MapMetrics();

// ---------------------------------------------------------------------------
// Seed data on startup
// ---------------------------------------------------------------------------
using (var scope = app.Services.CreateScope())
{
    var seeder = scope.ServiceProvider.GetRequiredService<CatalogSeeder>();
    await seeder.SeedAsync();
}

// ---------------------------------------------------------------------------
// Consul registration (optional, non-blocking)
// ---------------------------------------------------------------------------
if (builder.Configuration.GetSection("Consul").Exists())
{
    var lifetime = app.Lifetime;
    var consulClient = app.Services.GetRequiredService<IConsulClient>();
    var serviceName = "catalog-service";
    var serviceId = $"{serviceName}-{Guid.NewGuid():N}";
    var servicePort = builder.Configuration.GetValue<int>("Service:Port", 5100);
    var serviceHost = builder.Configuration.GetValue<string>("Service:Host") ?? "localhost";

    var registration = new AgentServiceRegistration
    {
        ID = serviceId,
        Name = serviceName,
        Address = serviceHost,
        Port = servicePort,
        Tags = new[] { "catalog", "api", "v1" },
        Check = new AgentServiceCheck
        {
            HTTP = $"http://{serviceHost}:{servicePort}/health/ready",
            Interval = TimeSpan.FromSeconds(10),
            Timeout = TimeSpan.FromSeconds(5),
            DeregisterCriticalServiceAfter = TimeSpan.FromSeconds(60)
        }
    };

    lifetime.ApplicationStarted.Register(async () =>
    {
        try
        {
            await consulClient.Agent.ServiceRegister(registration);
            app.Logger.LogInformation(
                "Registered service '{ServiceName}' (ID: {ServiceId}) with Consul at {Address}:{Port}.",
                serviceName, serviceId, serviceHost, servicePort);
        }
        catch (Exception ex)
        {
            app.Logger.LogWarning(ex, "Failed to register with Consul. Service discovery may not work.");
        }
    });

    lifetime.ApplicationStopping.Register(async () =>
    {
        app.Logger.LogInformation("Graceful shutdown initiated — deregistering from Consul...");
        try
        {
            await consulClient.Agent.ServiceDeregister(serviceId);
            app.Logger.LogInformation("Consul deregistration complete for '{ServiceId}'.", serviceId);
        }
        catch (Exception ex)
        {
            app.Logger.LogWarning(ex, "Failed to deregister from Consul.");
        }
        app.Logger.LogInformation("Waiting for background services to stop...");
    });

    lifetime.ApplicationStopped.Register(() =>
    {
        app.Logger.LogInformation("Shutdown complete. All resources released.");
    });
}

await app.RunAsync();
