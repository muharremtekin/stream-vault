using CatalogService.Application.Interfaces;
using CatalogService.Application.Validation;
using CatalogService.Domain.Entities;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Persistence.Repositories;

public class SeriesRepository : ISeriesRepository
{
    private readonly CatalogDbContext _context;

    public SeriesRepository(CatalogDbContext context)
    {
        _context = context ?? throw new ArgumentNullException(nameof(context));
    }

    public async Task<IReadOnlyList<Series>> GetAllAsync(
        int page = 1,
        int pageSize = 20,
        string? genre = null,
        string? sort = null,
        int? year = null,
        CancellationToken cancellationToken = default)
    {
        var filterBuilder = Builders<Series>.Filter;
        var filter = filterBuilder.Empty;

        if (!string.IsNullOrWhiteSpace(genre))
        {
            filter &= filterBuilder.AnyEq(s => s.Genres, genre);
        }

        if (year.HasValue)
        {
            filter &= filterBuilder.Eq(s => s.ReleaseYear, year.Value);
        }

        var sortDefinition = sort?.ToLowerInvariant() switch
        {
            "title" => Builders<Series>.Sort.Ascending(s => s.Title),
            "title_desc" => Builders<Series>.Sort.Descending(s => s.Title),
            "year" => Builders<Series>.Sort.Ascending(s => s.ReleaseYear),
            "year_desc" => Builders<Series>.Sort.Descending(s => s.ReleaseYear),
            "newest" => Builders<Series>.Sort.Descending(s => s.CreatedAt),
            _ => Builders<Series>.Sort.Descending(s => s.CreatedAt)
        };

        var seriesList = await _context.Series
            .Find(filter)
            .Sort(sortDefinition)
            .Skip((page - 1) * pageSize)
            .Limit(pageSize)
            .ToListAsync(cancellationToken);

        return seriesList.AsReadOnly();
    }

    public async Task<Series?> GetByIdAsync(string id, CancellationToken cancellationToken = default)
    {
        if (!CatalogObjectId.IsValid(id))
            return null;

        var filter = Builders<Series>.Filter.Eq(s => s.Id, id);
        return await _context.Series.Find(filter).FirstOrDefaultAsync(cancellationToken);
    }

    public async Task AddAsync(Series series, CancellationToken cancellationToken = default)
    {
        await _context.Series.InsertOneAsync(series, cancellationToken: cancellationToken);
    }

    public async Task UpdateAsync(Series series, CancellationToken cancellationToken = default)
    {
        var filter = Builders<Series>.Filter.Eq(s => s.Id, series.Id);
        await _context.Series.ReplaceOneAsync(filter, series, cancellationToken: cancellationToken);
    }

    public async Task<int> CountAsync(
        string? genre = null,
        int? year = null,
        CancellationToken cancellationToken = default)
    {
        var filterBuilder = Builders<Series>.Filter;
        var filter = filterBuilder.Empty;

        if (!string.IsNullOrWhiteSpace(genre))
        {
            filter &= filterBuilder.AnyEq(s => s.Genres, genre);
        }

        if (year.HasValue)
        {
            filter &= filterBuilder.Eq(s => s.ReleaseYear, year.Value);
        }

        var count = await _context.Series.CountDocumentsAsync(filter, cancellationToken: cancellationToken);
        return (int)count;
    }
}
