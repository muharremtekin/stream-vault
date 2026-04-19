using Microsoft.EntityFrameworkCore;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Repositories;

public class WatchlistRepository : IWatchlistRepository
{
    private readonly UserDbContext _context;

    public WatchlistRepository(UserDbContext context)
    {
        _context = context;
    }

    public async Task<List<WatchlistItem>> GetByProfileIdAsync(Guid profileId, CancellationToken cancellationToken = default)
    {
        return await _context.WatchlistItems
            .AsNoTracking()
            .Where(w => w.ProfileId == profileId)
            .OrderByDescending(w => w.AddedAt)
            .ToListAsync(cancellationToken);
    }

    public async Task AddAsync(WatchlistItem item, CancellationToken cancellationToken = default)
    {
        await _context.WatchlistItems.AddAsync(item, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task DeleteAsync(Guid id, CancellationToken cancellationToken = default)
    {
        await _context.WatchlistItems
            .Where(w => w.Id == id)
            .ExecuteDeleteAsync(cancellationToken);
    }
}
