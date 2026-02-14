using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface ISagaRepository
{
    Task<SagaState?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task AddAsync(SagaState saga, CancellationToken cancellationToken = default);
    Task UpdateAsync(SagaState saga, CancellationToken cancellationToken = default);
    Task<List<SagaState>> GetPendingSagasAsync(CancellationToken cancellationToken = default);
}
