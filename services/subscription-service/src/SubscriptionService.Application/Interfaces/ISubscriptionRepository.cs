using SubscriptionService.Application.DTOs;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface ISubscriptionRepository
{
    Task<Subscription?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task<Subscription?> GetActiveByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
    Task<Guid?> GetActiveSubscriptionIdByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
    Task<SubscriptionDto?> GetActiveDtoByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
    Task AddAsync(Subscription subscription, CancellationToken cancellationToken = default);
    Task UpdateAsync(Subscription subscription, CancellationToken cancellationToken = default);
    Task<List<ExpiredSubscriptionBatchItem>> GetExpiredSubscriptionsAsync(
        int batchSize,
        CancellationToken cancellationToken = default);
    Task UpdateRenewalBatchAsync(
        IReadOnlyCollection<SubscriptionRenewalUpdate> renewals,
        CancellationToken cancellationToken = default);
    Task<int> ExpireSubscriptionsAsync(
        IReadOnlyCollection<Guid> subscriptionIds,
        DateTime updatedAtUtc,
        CancellationToken cancellationToken = default);
}
