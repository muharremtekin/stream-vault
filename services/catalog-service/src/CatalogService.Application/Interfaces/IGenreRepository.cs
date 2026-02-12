using CatalogService.Domain.Entities;

namespace CatalogService.Application.Interfaces;

public interface IGenreRepository
{
    Task<IReadOnlyList<Genre>> GetAllAsync(CancellationToken cancellationToken = default);

    Task<Genre?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default);

    Task AddAsync(Genre genre, CancellationToken cancellationToken = default);

    Task AddManyAsync(IEnumerable<Genre> genres, CancellationToken cancellationToken = default);
}
