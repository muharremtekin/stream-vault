using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface IPaymentRepository
{
    Task<Payment?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task AddAsync(Payment payment, CancellationToken cancellationToken = default);
    Task UpdateAsync(Payment payment, CancellationToken cancellationToken = default);
    Task<List<Payment>> GetBySubscriptionIdAsync(Guid subscriptionId, CancellationToken cancellationToken = default);
    Task<List<Payment>> GetByUserIdAsync(Guid userId, int limit = 20, int offset = 0, CancellationToken cancellationToken = default);
}
