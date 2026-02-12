using UserService.Domain.Entities;

namespace UserService.Application.Interfaces;

public interface IWatchlistRepository
{
    Task<List<WatchlistItem>> GetByProfileIdAsync(Guid profileId, CancellationToken cancellationToken = default);

    Task AddAsync(WatchlistItem item, CancellationToken cancellationToken = default);

    Task DeleteAsync(Guid id, CancellationToken cancellationToken = default);
}
