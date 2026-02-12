using CatalogService.Domain.Entities;

namespace CatalogService.Application.Interfaces;

public interface IMovieRepository
{
    Task<IReadOnlyList<Movie>> GetAllAsync(
        int page = 1,
        int pageSize = 20,
        string? genre = null,
        string? sort = null,
        int? year = null,
        CancellationToken cancellationToken = default);

    Task<Movie?> GetByIdAsync(string id, CancellationToken cancellationToken = default);

    Task AddAsync(Movie movie, CancellationToken cancellationToken = default);

    Task UpdateAsync(Movie movie, CancellationToken cancellationToken = default);

    Task<int> CountAsync(
        string? genre = null,
        int? year = null,
        CancellationToken cancellationToken = default);
}
