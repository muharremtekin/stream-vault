using CatalogService.Domain.Entities;

namespace CatalogService.Application.Interfaces;

public interface IOutboxRepository
{
    Task AddAsync(OutboxMessage message, CancellationToken cancellationToken = default);
    Task<List<OutboxMessage>> GetUnprocessedAsync(int batchSize = 50, CancellationToken cancellationToken = default);
    Task MarkAsProcessedAsync(string id, CancellationToken cancellationToken = default);
    Task IncrementRetryAsync(string id, string errorMessage, CancellationToken cancellationToken = default);
}
