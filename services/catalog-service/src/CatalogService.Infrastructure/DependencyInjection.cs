using CatalogService.Application.Events;
using CatalogService.Application.Interfaces;
using CatalogService.Infrastructure.Messaging;
using CatalogService.Infrastructure.Persistence;
using CatalogService.Infrastructure.Persistence.Repositories;
using CatalogService.Infrastructure.Seed;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using MongoDB.Driver;

namespace CatalogService.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        // BsonClassMap configuration (must run before any MongoDB operations)
        BsonMappingConfiguration.Configure();

        // MongoDB configuration
        var connectionString = configuration.GetValue<string>("MongoDB:ConnectionString")
                               ?? "mongodb://localhost:27017";
        var databaseName = configuration.GetValue<string>("MongoDB:DatabaseName")
                           ?? Persistence.MongoCollectionSettings.DatabaseName;

        services.AddSingleton<IMongoClient>(_ => new MongoClient(connectionString));

        services.AddSingleton(sp =>
        {
            var client = sp.GetRequiredService<IMongoClient>();
            return client.GetDatabase(databaseName);
        });

        services.AddSingleton<CatalogDbContext>();

        // Repositories
        services.AddScoped<IMovieRepository, MovieRepository>();
        services.AddScoped<ISeriesRepository, SeriesRepository>();
        services.AddScoped<IGenreRepository, GenreRepository>();
        services.AddScoped<IOutboxRepository, OutboxRepository>();

        // Seeder
        services.AddTransient<CatalogSeeder>();

        // RabbitMQ
        services.AddSingleton<IEventPublisher, RabbitMqPublisher>();
        services.AddHostedService<EncodingResultConsumer>();
        services.AddHostedService<OutboxProcessor>();

        return services;
    }
}
