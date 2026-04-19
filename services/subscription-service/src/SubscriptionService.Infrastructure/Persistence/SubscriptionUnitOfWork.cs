using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Infrastructure.Persistence;

public class SubscriptionUnitOfWork : ISubscriptionUnitOfWork
{
    private readonly SubscriptionDbContext _context;

    public SubscriptionUnitOfWork(SubscriptionDbContext context)
    {
        _context = context;
    }

    public Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        return _context.SaveChangesAsync(cancellationToken);
    }

    public void DiscardPendingChanges()
    {
        _context.ChangeTracker.Clear();
    }
}
