using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class PaymentRepository : IPaymentRepository
{
    private readonly SubscriptionDbContext _context;

    public PaymentRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task<Payment?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Payments
            .FirstOrDefaultAsync(p => p.Id == id, cancellationToken);
    }

    public async Task AddAsync(Payment payment, CancellationToken cancellationToken = default)
    {
        await _context.Payments.AddAsync(payment, cancellationToken);
    }

    public Task UpdateAsync(Payment payment, CancellationToken cancellationToken = default)
    {
        var entry = _context.Entry(payment);
        if (entry.State == EntityState.Detached)
        {
            _context.Payments.Update(payment);
        }

        return Task.CompletedTask;
    }

    public async Task<List<Payment>> GetBySubscriptionIdAsync(Guid subscriptionId, CancellationToken cancellationToken = default)
    {
        return await _context.Payments
            .AsNoTracking()
            .Where(p => p.SubscriptionId == subscriptionId)
            .OrderByDescending(p => p.CreatedAt)
            .ToListAsync(cancellationToken);
    }

    public async Task<List<Payment>> GetByUserIdAsync(Guid userId, int limit = 20, int offset = 0, CancellationToken cancellationToken = default)
    {
        return await _context.Payments
            .AsNoTracking()
            .Where(p => p.Subscription.UserId == userId)
            .OrderByDescending(p => p.CreatedAt)
            .Skip(offset)
            .Take(limit)
            .ToListAsync(cancellationToken);
    }

    public async Task<List<PaymentDto>> GetHistoryByUserIdAsync(
        Guid userId,
        int limit = 20,
        int offset = 0,
        CancellationToken cancellationToken = default)
    {
        return await _context.Payments
            .AsNoTracking()
            .Where(p => p.Subscription.UserId == userId)
            .OrderByDescending(p => p.CreatedAt)
            .Skip(offset)
            .Take(limit)
            .Select(p => new PaymentDto
            {
                Id = p.Id,
                SubscriptionId = p.SubscriptionId,
                Amount = p.Amount,
                Currency = p.Currency,
                Status = p.Status.ToString(),
                TransactionId = p.TransactionId,
                CardLastFour = p.CardLastFour,
                CreatedAt = p.CreatedAt
            })
            .ToListAsync(cancellationToken);
    }
}
