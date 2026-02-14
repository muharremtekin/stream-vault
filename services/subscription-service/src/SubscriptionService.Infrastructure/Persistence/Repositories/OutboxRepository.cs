using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class OutboxRepository : IOutboxRepository
{
    private readonly SubscriptionDbContext _context;

    public OutboxRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task AddAsync(OutboxMessage message, CancellationToken cancellationToken = default)
    {
        await _context.OutboxMessages.AddAsync(message, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task<List<OutboxMessage>> GetUnprocessedAsync(int batchSize = 50, CancellationToken cancellationToken = default)
    {
        return await _context.OutboxMessages
            .Where(o => o.ProcessedAt == null && o.RetryCount < 3)
            .OrderBy(o => o.CreatedAt)
            .Take(batchSize)
            .ToListAsync(cancellationToken);
    }

    public async Task MarkAsProcessedAsync(Guid id, CancellationToken cancellationToken = default)
    {
        var message = await _context.OutboxMessages.FindAsync(new object[] { id }, cancellationToken);
        if (message is not null)
        {
            message.ProcessedAt = DateTime.UtcNow;
            await _context.SaveChangesAsync(cancellationToken);
        }
    }

    public async Task IncrementRetryAsync(Guid id, string errorMessage, CancellationToken cancellationToken = default)
    {
        var message = await _context.OutboxMessages.FindAsync(new object[] { id }, cancellationToken);
        if (message is not null)
        {
            message.RetryCount++;
            message.ErrorMessage = errorMessage;
            await _context.SaveChangesAsync(cancellationToken);
        }
    }
}
