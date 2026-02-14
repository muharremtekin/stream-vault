using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface ISubscriptionRepository
{
    Task<Subscription?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task<Subscription?> GetActiveByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
    Task AddAsync(Subscription subscription, CancellationToken cancellationToken = default);
    Task UpdateAsync(Subscription subscription, CancellationToken cancellationToken = default);
    Task<List<Subscription>> GetExpiredSubscriptionsAsync(CancellationToken cancellationToken = default);
}
