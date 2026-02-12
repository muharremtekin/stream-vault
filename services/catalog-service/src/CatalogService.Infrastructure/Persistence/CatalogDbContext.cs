using CatalogService.Domain.Entities;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Persistence;

public class CatalogDbContext
{
    private readonly IMongoDatabase _database;

    public CatalogDbContext(IMongoDatabase database)
    {
        _database = database ?? throw new ArgumentNullException(nameof(database));
    }

    public IMongoCollection<Movie> Movies =>
        _database.GetCollection<Movie>(MongoCollectionSettings.MoviesCollection);

    public IMongoCollection<Series> Series =>
        _database.GetCollection<Series>(MongoCollectionSettings.SeriesCollection);

    public IMongoCollection<Genre> Genres =>
        _database.GetCollection<Genre>(MongoCollectionSettings.GenresCollection);

    public async Task CreateIndexesAsync()
    {
        // Movie indexes
        var movieIndexes = Builders<Movie>.IndexKeys;
        await Movies.Indexes.CreateManyAsync(new[]
        {
            new CreateIndexModel<Movie>(movieIndexes.Ascending(m => m.Title)),
            new CreateIndexModel<Movie>(movieIndexes.Ascending(m => m.Genres)),
            new CreateIndexModel<Movie>(movieIndexes.Ascending(m => m.ReleaseYear)),
            new CreateIndexModel<Movie>(movieIndexes.Ascending(m => m.Status)),
            new CreateIndexModel<Movie>(movieIndexes.Descending(m => m.AverageRating))
        });

        // Series indexes
        var seriesIndexes = Builders<Series>.IndexKeys;
        await Series.Indexes.CreateManyAsync(new[]
        {
            new CreateIndexModel<Series>(seriesIndexes.Ascending(s => s.Title)),
            new CreateIndexModel<Series>(seriesIndexes.Ascending(s => s.Genres)),
            new CreateIndexModel<Series>(seriesIndexes.Ascending(s => s.ReleaseYear)),
            new CreateIndexModel<Series>(seriesIndexes.Ascending(s => s.Status))
        });

        // Genre indexes
        var genreIndexes = Builders<Genre>.IndexKeys;
        await Genres.Indexes.CreateManyAsync(new[]
        {
            new CreateIndexModel<Genre>(
                genreIndexes.Ascending(g => g.Slug),
                new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Genre>(genreIndexes.Ascending(g => g.Name))
        });
    }
}
