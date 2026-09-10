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
        if (!message.NextAttemptAt.HasValue)
        {
            message.NextAttemptAt = message.CreatedAt == default
                ? DateTime.UtcNow
                : message.CreatedAt;
        }

        await _context.OutboxMessages.AddAsync(message, cancellationToken);
    }

    public async Task<List<OutboxMessage>> GetUnprocessedAsync(int batchSize = 50, CancellationToken cancellationToken = default)
    {
        var now = DateTime.UtcNow;

        return await _context.OutboxMessages
            .Where(o => o.ProcessedAt == null
                && !o.IsDeadLetter
                && ((o.NextAttemptAt ?? o.CreatedAt) <= now))
            .OrderBy(o => o.NextAttemptAt ?? o.CreatedAt)
            .ThenBy(o => o.CreatedAt)
            .Take(batchSize)
            .ToListAsync(cancellationToken);
    }

    public async Task MarkAsProcessedAsync(Guid id, CancellationToken cancellationToken = default)
    {
        await _context.OutboxMessages
            .Where(o => o.Id == id)
            .ExecuteUpdateAsync(
                s => s
                    .SetProperty(o => o.ProcessedAt, DateTime.UtcNow)
                    .SetProperty(o => o.NextAttemptAt, (DateTime?)null),
                cancellationToken);
    }

    public async Task IncrementRetryAsync(Guid id, string errorMessage, CancellationToken cancellationToken = default)
    {
        var retryState = await _context.OutboxMessages
            .Where(o => o.Id == id)
            .Select(o => new { o.RetryCount })
            .SingleAsync(cancellationToken);

        var attemptTime = DateTime.UtcNow;
        var nextRetryCount = retryState.RetryCount + 1;
        var backoffDelay = TimeSpan.FromSeconds(Math.Min(Math.Pow(2, nextRetryCount), 60));
        var nextAttemptAt = attemptTime.Add(backoffDelay);

        await _context.OutboxMessages
            .Where(o => o.Id == id)
            .ExecuteUpdateAsync(
                s => s
                    .SetProperty(o => o.RetryCount, o => o.RetryCount + 1)
                    .SetProperty(o => o.ErrorMessage, errorMessage)
                    .SetProperty(o => o.LastAttemptedAt, attemptTime)
                    .SetProperty(o => o.NextAttemptAt, (DateTime?)nextAttemptAt),
                cancellationToken);
    }

    public async Task MarkAsDeadLetterAsync(Guid id, CancellationToken cancellationToken = default)
    {
        await _context.OutboxMessages
            .Where(o => o.Id == id)
            .ExecuteUpdateAsync(
                s => s
                    .SetProperty(o => o.IsDeadLetter, true)
                    .SetProperty(o => o.LastAttemptedAt, DateTime.UtcNow)
                    .SetProperty(o => o.NextAttemptAt, (DateTime?)null),
                cancellationToken);
    }

    public async Task<List<OutboxMessage>> GetDeadLetterMessagesAsync(int batchSize = 50, CancellationToken cancellationToken = default)
    {
        return await _context.OutboxMessages
            .Where(o => o.IsDeadLetter)
            .OrderBy(o => o.CreatedAt)
            .Take(batchSize)
            .ToListAsync(cancellationToken);
    }
}
