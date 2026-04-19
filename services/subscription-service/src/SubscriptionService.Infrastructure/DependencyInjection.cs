using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Application.Sagas;
using SubscriptionService.Infrastructure.Services;
using SubscriptionService.Infrastructure.Persistence;
using SubscriptionService.Infrastructure.Persistence.Repositories;

namespace SubscriptionService.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        services.AddDbContextPool<SubscriptionDbContext>(options =>
            options.UseNpgsql(
                configuration.GetConnectionString("DefaultConnection"),
                npgsqlOptions =>
                {
                    npgsqlOptions.MigrationsAssembly(typeof(SubscriptionDbContext).Assembly.FullName);
                    npgsqlOptions.EnableRetryOnFailure(
                        maxRetryCount: 3,
                        maxRetryDelay: TimeSpan.FromSeconds(10),
                        errorCodesToAdd: null);
                }));

        services.AddScoped<IPlanRepository, PlanRepository>();
        services.AddScoped<ISubscriptionRepository, SubscriptionRepository>();
        services.AddScoped<IPaymentRepository, PaymentRepository>();
        services.AddScoped<ISagaRepository, SagaRepository>();
        services.AddScoped<IOutboxRepository, OutboxRepository>();
        services.AddScoped<IInvoiceRepository, InvoiceRepository>();
        services.AddScoped<ISubscriptionUnitOfWork, SubscriptionUnitOfWork>();
        services.AddScoped<IPaymentGateway, MockPaymentGateway>();

        services.AddSingleton<IRabbitMqPublisher, RabbitMqPublisher>();

        services.AddScoped<SubscriptionSaga>();
        services.AddScoped<ChangePlanSaga>();
        services.AddScoped<CompensatingActions>();

        services.AddScoped<DbSeeder>();

        return services;
    }
}
