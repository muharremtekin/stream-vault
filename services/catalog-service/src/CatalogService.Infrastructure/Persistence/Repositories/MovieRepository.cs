using CatalogService.Application.Interfaces;
using CatalogService.Application.Validation;
using CatalogService.Domain.Entities;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Persistence.Repositories;

public class MovieRepository : IMovieRepository
{
    private readonly CatalogDbContext _context;

    public MovieRepository(CatalogDbContext context)
    {
        _context = context ?? throw new ArgumentNullException(nameof(context));
    }

    public async Task<IReadOnlyList<Movie>> GetAllAsync(
        int page = 1,
        int pageSize = 20,
        string? genre = null,
        string? sort = null,
        int? year = null,
        CancellationToken cancellationToken = default)
    {
        var filterBuilder = Builders<Movie>.Filter;
        var filter = filterBuilder.Empty;

        if (!string.IsNullOrWhiteSpace(genre))
        {
            filter &= filterBuilder.AnyEq(m => m.Genres, genre);
        }

        if (year.HasValue)
        {
            filter &= filterBuilder.Eq(m => m.ReleaseYear, year.Value);
        }

        var sortDefinition = sort?.ToLowerInvariant() switch
        {
            "title" => Builders<Movie>.Sort.Ascending(m => m.Title),
            "title_desc" => Builders<Movie>.Sort.Descending(m => m.Title),
            "year" => Builders<Movie>.Sort.Ascending(m => m.ReleaseYear),
            "year_desc" => Builders<Movie>.Sort.Descending(m => m.ReleaseYear),
            "rating" => Builders<Movie>.Sort.Descending(m => m.AverageRating),
            "newest" => Builders<Movie>.Sort.Descending(m => m.CreatedAt),
            _ => Builders<Movie>.Sort.Descending(m => m.CreatedAt)
        };

        var movies = await _context.Movies
            .Find(filter)
            .Sort(sortDefinition)
            .Skip((page - 1) * pageSize)
            .Limit(pageSize)
            .ToListAsync(cancellationToken);

        return movies.AsReadOnly();
    }

    public async Task<Movie?> GetByIdAsync(string id, CancellationToken cancellationToken = default)
    {
        if (!CatalogObjectId.IsValid(id))
            return null;

        var filter = Builders<Movie>.Filter.Eq(m => m.Id, id);
        return await _context.Movies.Find(filter).FirstOrDefaultAsync(cancellationToken);
    }

    public async Task AddAsync(Movie movie, CancellationToken cancellationToken = default)
    {
        await _context.Movies.InsertOneAsync(movie, cancellationToken: cancellationToken);
    }

    public async Task UpdateAsync(Movie movie, CancellationToken cancellationToken = default)
    {
        var filter = Builders<Movie>.Filter.Eq(m => m.Id, movie.Id);
        await _context.Movies.ReplaceOneAsync(filter, movie, cancellationToken: cancellationToken);
    }

    public async Task<int> CountAsync(
        string? genre = null,
        int? year = null,
        CancellationToken cancellationToken = default)
    {
        var filterBuilder = Builders<Movie>.Filter;
        var filter = filterBuilder.Empty;

        if (!string.IsNullOrWhiteSpace(genre))
        {
            filter &= filterBuilder.AnyEq(m => m.Genres, genre);
        }

        if (year.HasValue)
        {
            filter &= filterBuilder.Eq(m => m.ReleaseYear, year.Value);
        }

        var count = await _context.Movies.CountDocumentsAsync(filter, cancellationToken: cancellationToken);
        return (int)count;
    }
}
