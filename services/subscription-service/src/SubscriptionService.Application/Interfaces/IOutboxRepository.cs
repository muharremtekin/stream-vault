using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface IOutboxRepository
{
    Task AddAsync(OutboxMessage message, CancellationToken cancellationToken = default);
    Task<List<OutboxMessage>> GetUnprocessedAsync(int batchSize = 50, CancellationToken cancellationToken = default);
    Task MarkAsProcessedAsync(Guid id, CancellationToken cancellationToken = default);
    Task IncrementRetryAsync(Guid id, string errorMessage, CancellationToken cancellationToken = default);
    Task MarkAsDeadLetterAsync(Guid id, CancellationToken cancellationToken = default);
    Task<List<OutboxMessage>> GetDeadLetterMessagesAsync(int batchSize = 50, CancellationToken cancellationToken = default);
}
