using Consul;
using FluentValidation;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.EntityFrameworkCore;
using Prometheus;
using Microsoft.OpenApi.Models;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using Serilog;
using Serilog.Formatting.Compact;
using SubscriptionService.Api.BackgroundServices;
using SubscriptionService.Api.Filters;
using SubscriptionService.Api.Middleware;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Infrastructure;
using SubscriptionService.Infrastructure.Persistence;

var builder = WebApplication.CreateBuilder(args);

// Serilog
builder.Host.UseSerilog((context, loggerConfig) =>
{
    loggerConfig
        .ReadFrom.Configuration(context.Configuration)
        .Enrich.FromLogContext()
        .Enrich.WithProperty("ServiceName", "subscription-service")
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
                    .AddService("subscription-service", serviceVersion: "1.0.0"))
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddEntityFrameworkCoreInstrumentation()
                .AddOtlpExporter(opts =>
                {
                    opts.Endpoint = new Uri(otelEndpoint);
                });
        });
}

// -------------------------------------------------------------------
// Services
// -------------------------------------------------------------------

// Controllers
builder.Services.AddControllers(options =>
{
    options.Filters.Add<ValidationFilter>();
});

// Swagger / OpenAPI
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(options =>
{
    options.SwaggerDoc("v1", new OpenApiInfo
    {
        Title = "StreamVault Subscription Service",
        Version = "v1",
        Description = "Subscription management, plans, payments, and billing for StreamVault."
    });
});

// MediatR
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssembly(typeof(IPlanRepository).Assembly));

// FluentValidation
builder.Services.AddValidatorsFromAssembly(typeof(IPlanRepository).Assembly);

// AutoMapper
builder.Services.AddAutoMapper(typeof(IPlanRepository).Assembly);

// Infrastructure (EF Core, Repositories)
builder.Services.AddInfrastructure(builder.Configuration);

// Health checks
builder.Services.AddHealthChecks()
    .AddNpgSql(
        builder.Configuration.GetConnectionString("DefaultConnection")!,
        name: "postgresql",
        tags: new[] { "ready" })
    .AddRabbitMQ(
        new Uri(builder.Configuration["RabbitMQ:ConnectionString"] ?? "amqp://guest:guest@localhost:5672/"),
        name: "rabbitmq",
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

// Background services
builder.Services.AddHostedService<OutboxProcessorService>();
builder.Services.AddHostedService<SubscriptionRenewalService>();

// Consul service discovery
builder.Services.AddSingleton<IConsulClient, ConsulClient>(_ =>
    new ConsulClient(config =>
    {
        var consulAddress = builder.Configuration.GetValue<string>("Consul:Address")
            ?? builder.Configuration.GetValue<string>("Consul:Host")
            ?? "http://localhost:8500";
        config.Address = new Uri(consulAddress);
    }));

// -------------------------------------------------------------------
// App pipeline
// -------------------------------------------------------------------

var app = builder.Build();

// Auto-apply EF Core migrations and seed data (skip for testing environments)
if (!app.Environment.IsEnvironment("Testing"))
{
    using var scope = app.Services.CreateScope();
    var dbContext = scope.ServiceProvider.GetRequiredService<SubscriptionDbContext>();
    try
    {
        app.Logger.LogInformation("Applying database migrations...");
        await dbContext.Database.MigrateAsync();
        app.Logger.LogInformation("Database migrations applied successfully.");

        var seeder = scope.ServiceProvider.GetRequiredService<DbSeeder>();
        await seeder.SeedAsync();
    }
    catch (Exception ex)
    {
        app.Logger.LogError(ex, "An error occurred while applying database migrations.");
        throw;
    }
}

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI(options =>
    {
        options.SwaggerEndpoint("/swagger/v1/swagger.json", "StreamVault Subscription Service v1");
    });
}

app.UseCorrelationId();
app.UseSerilogRequestLogging();
app.UseExceptionHandling();
app.UseHttpMetrics();

app.UseCors("AllowAll");

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

// -------------------------------------------------------------------
// Consul service registration (optional, non-blocking)
// -------------------------------------------------------------------
if (builder.Configuration.GetSection("Consul").Exists())
{
    var lifetime = app.Lifetime;
    var consulClient = app.Services.GetRequiredService<IConsulClient>();

    var serviceName = builder.Configuration.GetValue<string>("Consul:ServiceName")
        ?? builder.Configuration.GetValue<string>("ServiceRegistration:Name")
        ?? "subscription-service";
    var serviceId = $"{serviceName}-{Guid.NewGuid():N}";
    var servicePort = builder.Configuration.GetValue<int>("Consul:ServicePort",
        builder.Configuration.GetValue<int>("ServiceRegistration:Port", 5007));
    var serviceHost = builder.Configuration.GetValue<string>("Service:Host") ?? "localhost";

    var registration = new AgentServiceRegistration
    {
        ID = serviceId,
        Name = serviceName,
        Address = serviceHost,
        Port = servicePort,
        Tags = new[] { "subscription", "billing", "api", "v1" },
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
        try
        {
            await consulClient.Agent.ServiceDeregister(serviceId);
            app.Logger.LogInformation("Deregistered service '{ServiceId}' from Consul.", serviceId);
        }
        catch (Exception ex)
        {
            app.Logger.LogWarning(ex, "Failed to deregister from Consul.");
        }
    });
}

await app.RunAsync();

// Expose partial class for integration tests with WebApplicationFactory
public partial class Program { }
