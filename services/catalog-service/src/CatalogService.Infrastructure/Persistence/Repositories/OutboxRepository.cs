using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using MongoDB.Driver;

namespace CatalogService.Infrastructure.Persistence.Repositories;

public class OutboxRepository : IOutboxRepository
{
    private readonly CatalogDbContext _context;

    public OutboxRepository(CatalogDbContext context)
    {
        _context = context;
    }

    public async Task AddAsync(OutboxMessage message, CancellationToken cancellationToken = default)
    {
        await _context.OutboxMessages.InsertOneAsync(message, cancellationToken: cancellationToken);
    }

    public async Task<List<OutboxMessage>> GetUnprocessedAsync(int batchSize = 50, CancellationToken cancellationToken = default)
    {
        var filter = Builders<OutboxMessage>.Filter.And(
            Builders<OutboxMessage>.Filter.Eq(m => m.ProcessedAt, null),
            Builders<OutboxMessage>.Filter.Lt(m => m.RetryCount, 3)
        );

        return await _context.OutboxMessages
            .Find(filter)
            .SortBy(m => m.CreatedAt)
            .Limit(batchSize)
            .ToListAsync(cancellationToken);
    }

    public async Task MarkAsProcessedAsync(string id, CancellationToken cancellationToken = default)
    {
        var filter = Builders<OutboxMessage>.Filter.Eq(m => m.Id, id);
        var update = Builders<OutboxMessage>.Update.Set(m => m.ProcessedAt, DateTime.UtcNow);

        await _context.OutboxMessages.UpdateOneAsync(filter, update, cancellationToken: cancellationToken);
    }

    public async Task IncrementRetryAsync(string id, string errorMessage, CancellationToken cancellationToken = default)
    {
        var filter = Builders<OutboxMessage>.Filter.Eq(m => m.Id, id);
        var update = Builders<OutboxMessage>.Update
            .Inc(m => m.RetryCount, 1)
            .Set(m => m.ErrorMessage, errorMessage);

        await _context.OutboxMessages.UpdateOneAsync(filter, update, cancellationToken: cancellationToken);
    }
}
