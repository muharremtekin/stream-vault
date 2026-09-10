using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class SagaRepository : ISagaRepository
{
    private readonly SubscriptionDbContext _context;

    public SagaRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task<SagaState?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.SagaStates
            .FirstOrDefaultAsync(s => s.Id == id, cancellationToken);
    }

    public async Task AddAsync(SagaState saga, CancellationToken cancellationToken = default)
    {
        await _context.SagaStates.AddAsync(saga, cancellationToken);
    }

    public Task UpdateAsync(SagaState saga, CancellationToken cancellationToken = default)
    {
        saga.UpdatedAt = DateTime.UtcNow;

        var entry = _context.Entry(saga);
        if (entry.State == EntityState.Detached)
        {
            _context.SagaStates.Update(saga);
            return Task.CompletedTask;
        }

        entry.Property(s => s.UpdatedAt).IsModified = true;
        return Task.CompletedTask;
    }

    public async Task<List<SagaState>> GetPendingSagasAsync(CancellationToken cancellationToken = default)
    {
        return await _context.SagaStates
            .Where(s => s.Status == SagaStatus.Started || s.Status == SagaStatus.InProgress)
            .OrderBy(s => s.CreatedAt)
            .ToListAsync(cancellationToken);
    }
}
