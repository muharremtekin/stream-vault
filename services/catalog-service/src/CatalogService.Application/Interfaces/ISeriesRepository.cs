using CatalogService.Domain.Entities;

namespace CatalogService.Application.Interfaces;

public interface ISeriesRepository
{
    Task<IReadOnlyList<Series>> GetAllAsync(
        int page = 1,
        int pageSize = 20,
        string? genre = null,
        string? sort = null,
        int? year = null,
        CancellationToken cancellationToken = default);

    Task<Series?> GetByIdAsync(string id, CancellationToken cancellationToken = default);

    Task AddAsync(Series series, CancellationToken cancellationToken = default);

    Task UpdateAsync(Series series, CancellationToken cancellationToken = default);

    Task<int> CountAsync(
        string? genre = null,
        int? year = null,
        CancellationToken cancellationToken = default);
}
