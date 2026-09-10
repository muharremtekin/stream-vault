using System.Text;
using System.Text.Json.Serialization;
using Consul;
using FluentValidation;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using Microsoft.OpenApi.Models;
using UserService.Api.Filters;
using UserService.Api.Middleware;
using UserService.Application.Commands.RegisterUser;
using UserService.Application.Mappings;
using UserService.Infrastructure;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using Serilog;
using Serilog.Formatting.Compact;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Prometheus;
using UserService.Infrastructure.Persistence;
using UserService.Infrastructure.Discovery;
using UserService.Api.Health;

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
        .Enrich.WithProperty("ServiceName", "user-service")
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
                    .AddService("user-service", serviceVersion: "1.0.0"))
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddEntityFrameworkCoreInstrumentation()
                .AddSource("UserService")
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
}).AddJsonOptions(options =>
{
    options.JsonSerializerOptions.Converters.Add(
        new JsonStringEnumConverter(System.Text.Json.JsonNamingPolicy.CamelCase));
});

// Swagger / OpenAPI
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(options =>
{
    options.SwaggerDoc("v1", new OpenApiInfo
    {
        Title = "StreamVault User Service",
        Version = "v1",
        Description = "User management, authentication, profiles, and watchlists for StreamVault."
    });

    options.AddSecurityDefinition("Bearer", new OpenApiSecurityScheme
    {
        Name = "Authorization",
        Type = SecuritySchemeType.Http,
        Scheme = "bearer",
        BearerFormat = "JWT",
        In = ParameterLocation.Header,
        Description = "Enter your JWT token."
    });

    options.AddSecurityRequirement(new OpenApiSecurityRequirement
    {
        {
            new OpenApiSecurityScheme
            {
                Reference = new OpenApiReference
                {
                    Type = ReferenceType.SecurityScheme,
                    Id = "Bearer"
                }
            },
            Array.Empty<string>()
        }
    });
});

// MediatR
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssembly(typeof(RegisterUserCommand).Assembly));

// FluentValidation
builder.Services.AddValidatorsFromAssembly(typeof(RegisterUserValidator).Assembly);

// AutoMapper
builder.Services.AddAutoMapper(typeof(MappingProfile).Assembly);

// Infrastructure (EF Core, Repositories, Services)
builder.Services.AddInfrastructure(builder.Configuration);

// JWT Authentication
var jwtSecret = builder.Configuration["Jwt:Secret"]
    ?? throw new InvalidOperationException("JWT secret is not configured.");

builder.Services.AddAuthentication(options =>
{
    options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
    options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
})
.AddJwtBearer(options =>
{
    options.TokenValidationParameters = new TokenValidationParameters
    {
        ValidateIssuer = true,
        ValidateAudience = true,
        ValidateLifetime = true,
        ValidateIssuerSigningKey = true,
        ValidIssuer = builder.Configuration["Jwt:Issuer"],
        ValidAudience = builder.Configuration["Jwt:Audience"],
        IssuerSigningKey = new SymmetricSecurityKey(
            Encoding.UTF8.GetBytes(jwtSecret)),
        ClockSkew = TimeSpan.Zero
    };
});

builder.Services.AddAuthorization();

// Health checks
builder.Services.AddHealthChecks()
    .AddNpgSql(
        builder.Configuration.GetConnectionString("DefaultConnection")!,
        name: "postgresql",
        tags: new[] { "ready" })
	.AddCheck<RabbitMqHealthCheck>(
		name: "rabbitmq",
		tags: new[] { "ready" });

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
    var dbContext = scope.ServiceProvider.GetRequiredService<UserDbContext>();
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
        options.SwaggerEndpoint("/swagger/v1/swagger.json", "StreamVault User Service v1");
    });
}

app.UseCorrelationId();
app.UseSerilogRequestLogging();
app.UseExceptionHandling();
app.UseHttpMetrics();

app.UseAuthentication();
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

// -------------------------------------------------------------------
// Consul service registration (optional, non-blocking)
// -------------------------------------------------------------------
if (builder.Configuration.GetSection("Consul").Exists())
{
    var lifetime = app.Lifetime;
    var consulClient = app.Services.GetRequiredService<IConsulClient>();

    var registrationSettings = ConsulRegistrationSettings.FromConfiguration(
        builder.Configuration, "user-service", 8080);
    var serviceName = registrationSettings.Name;
    var serviceId = $"{serviceName}-{Guid.NewGuid():N}";
    var registration = new AgentServiceRegistration
    {
        ID = serviceId,
        Name = registrationSettings.Name,
        Address = registrationSettings.Host,
        Port = registrationSettings.Port,
        Tags = new[] { "user", "auth", "api", "v1" },
        Check = new AgentServiceCheck
        {
            HTTP = registrationSettings.HealthCheckAddress,
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
                serviceName, serviceId, registrationSettings.Host, registrationSettings.Port);
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

// Expose partial class for integration tests with WebApplicationFactory
public partial class Program { }
