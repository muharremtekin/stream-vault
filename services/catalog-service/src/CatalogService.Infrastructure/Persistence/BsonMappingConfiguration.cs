using CatalogService.Domain.ValueObjects;
using MongoDB.Bson.Serialization;

namespace CatalogService.Infrastructure.Persistence;

public static class BsonMappingConfiguration
{
    private static bool _configured;

    public static void Configure()
    {
        if (_configured) return;

        BsonClassMap.RegisterClassMap<Duration>(cm =>
        {
            cm.AutoMap();
            cm.UnmapMember(d => d.TotalMinutes);
        });

        _configured = true;
    }
}
