using CatalogService.Api.Middleware;
using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Application.Mappings;
using CatalogService.Infrastructure;
using CatalogService.Infrastructure.Seed;
using Consul;
using FluentValidation;
using FluentValidation.AspNetCore;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Serilog;
using Serilog.Formatting.Compact;

var builder = WebApplication.CreateBuilder(args);

// Serilog
builder.Host.UseSerilog((context, loggerConfig) =>
{
    loggerConfig
        .ReadFrom.Configuration(context.Configuration)
        .Enrich.FromLogContext()
        .WriteTo.Console(new RenderedCompactJsonFormatter());
});

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
app.MapHealthChecks("/health", new HealthCheckOptions { Predicate = _ => false });

// ---------------------------------------------------------------------------
// Seed data on startup
// ---------------------------------------------------------------------------
using (var scope = app.Services.CreateScope())
{
    var seeder = scope.ServiceProvider.GetRequiredService<CatalogSeeder>();
    await seeder.SeedAsync();
}

// ---------------------------------------------------------------------------
// Consul registration
// ---------------------------------------------------------------------------
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
        HTTP = $"http://{serviceHost}:{servicePort}/health/live",
        Interval = TimeSpan.FromSeconds(10),
        Timeout = TimeSpan.FromSeconds(5),
        DeregisterCriticalServiceAfter = TimeSpan.FromSeconds(30)
    }
};

lifetime.ApplicationStarted.Register(async () =>
{
    try
    {
        await consulClient.Agent.ServiceRegister(registration);
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
    }
    catch (Exception ex)
    {
        app.Logger.LogWarning(ex, "Failed to deregister from Consul.");
    }
});

await app.RunAsync();
