using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Persistence.Repositories;

public class GenreRepository : IGenreRepository
{
    private readonly CatalogDbContext _context;

    public GenreRepository(CatalogDbContext context)
    {
        _context = context ?? throw new ArgumentNullException(nameof(context));
    }

    public async Task<IReadOnlyList<Genre>> GetAllAsync(CancellationToken cancellationToken = default)
    {
        var genres = await _context.Genres
            .Find(Builders<Genre>.Filter.Empty)
            .SortBy(g => g.Name)
            .ToListAsync(cancellationToken);

        return genres.AsReadOnly();
    }

    public async Task<Genre?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default)
    {
        var filter = Builders<Genre>.Filter.Eq(g => g.Slug, slug);
        return await _context.Genres.Find(filter).FirstOrDefaultAsync(cancellationToken);
    }

    public async Task AddAsync(Genre genre, CancellationToken cancellationToken = default)
    {
        await _context.Genres.InsertOneAsync(genre, cancellationToken: cancellationToken);
    }

    public async Task AddManyAsync(IEnumerable<Genre> genres, CancellationToken cancellationToken = default)
    {
        var genreList = genres.ToList();
        if (genreList.Count > 0)
        {
            await _context.Genres.InsertManyAsync(genreList, cancellationToken: cancellationToken);
        }
    }
}
