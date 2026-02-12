using CatalogService.Domain.ValueObjects;
using MongoDB.Bson.Serialization;
using MongoDB.Bson.Serialization.Conventions;

namespace CatalogService.Infrastructure.Persistence;

public static class BsonMappingConfiguration
{
    private static bool _configured;

    public static void Configure()
    {
        if (_configured) return;

        // Register camelCase convention so BSON fields match MongoDB indexes
        var conventionPack = new ConventionPack
        {
            new CamelCaseElementNameConvention()
        };
        ConventionRegistry.Register("camelCase", conventionPack, _ => true);

        BsonClassMap.RegisterClassMap<Duration>(cm =>
        {
            cm.AutoMap();
            cm.UnmapMember(d => d.TotalMinutes);
        });

        _configured = true;
    }
}
