namespace CatalogService.Infrastructure.Persistence;

public static class MongoCollectionSettings
{
    public const string MoviesCollection = "movies";
    public const string SeriesCollection = "series";
    public const string GenresCollection = "genres";
    public const string OutboxCollection = "outbox_messages";

    public const string DatabaseName = "stream_vault_catalog";
}
