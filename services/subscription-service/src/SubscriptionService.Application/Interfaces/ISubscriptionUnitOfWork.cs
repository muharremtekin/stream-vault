namespace SubscriptionService.Application.Interfaces;

public interface ISubscriptionUnitOfWork
{
    Task<int> SaveChangesAsync(CancellationToken cancellationToken = default);
}
