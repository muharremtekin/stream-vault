using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class SubscriptionRepository : ISubscriptionRepository
{
    private readonly SubscriptionDbContext _context;

    public SubscriptionRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task<Subscription?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Subscriptions
            .Include(s => s.Plan)
            .FirstOrDefaultAsync(s => s.Id == id, cancellationToken);
    }

    public async Task<Subscription?> GetActiveByUserIdAsync(Guid userId, CancellationToken cancellationToken = default)
    {
        return await _context.Subscriptions
            .Include(s => s.Plan)
            .FirstOrDefaultAsync(
                s => s.UserId == userId && s.Status == SubscriptionStatus.Active,
                cancellationToken);
    }

    public async Task<Guid?> GetActiveSubscriptionIdByUserIdAsync(
        Guid userId,
        CancellationToken cancellationToken = default)
    {
        return await _context.Subscriptions
            .AsNoTracking()
            .Where(s => s.UserId == userId && s.Status == SubscriptionStatus.Active)
            .Select(s => (Guid?)s.Id)
            .FirstOrDefaultAsync(cancellationToken);
    }

    public async Task<SubscriptionDto?> GetActiveDtoByUserIdAsync(
        Guid userId,
        CancellationToken cancellationToken = default)
    {
        return await _context.Subscriptions
            .AsNoTracking()
            .Where(s => s.UserId == userId && s.Status == SubscriptionStatus.Active)
            .Select(s => new SubscriptionDto
            {
                Id = s.Id,
                UserId = s.UserId,
                Status = s.Status.ToString(),
                PeriodStart = s.PeriodStart,
                PeriodEnd = s.PeriodEnd,
                AutoRenew = s.AutoRenew,
                CancelledAt = s.CancelledAt,
                CreatedAt = s.CreatedAt,
                Plan = new PlanDto
                {
                    Id = s.Plan.Id,
                    Name = s.Plan.Name,
                    Tier = s.Plan.Tier.ToString(),
                    PriceMonthly = s.Plan.PriceMonthly,
                    MaxScreens = s.Plan.MaxScreens,
                    MaxQuality = s.Plan.MaxQuality,
                    Features = s.Plan.Features,
                    IsActive = s.Plan.IsActive
                }
            })
            .FirstOrDefaultAsync(cancellationToken);
    }

    public async Task AddAsync(Subscription subscription, CancellationToken cancellationToken = default)
    {
        await _context.Subscriptions.AddAsync(subscription, cancellationToken);
    }

    public Task UpdateAsync(Subscription subscription, CancellationToken cancellationToken = default)
    {
        subscription.UpdatedAt = DateTime.UtcNow;

        var entry = _context.Entry(subscription);
        if (entry.State == EntityState.Detached)
        {
            _context.Subscriptions.Update(subscription);
            return Task.CompletedTask;
        }

        entry.Property(s => s.UpdatedAt).IsModified = true;
        return Task.CompletedTask;
    }

    public async Task<List<ExpiredSubscriptionBatchItem>> GetExpiredSubscriptionsAsync(
        int batchSize,
        CancellationToken cancellationToken = default)
    {
        var now = DateTime.UtcNow;

        return await _context.Subscriptions
            .AsNoTracking()
            .Where(s => s.Status == SubscriptionStatus.Active && s.PeriodEnd <= now)
            .OrderBy(s => s.PeriodEnd)
            .ThenBy(s => s.Id)
            .Select(s => new ExpiredSubscriptionBatchItem
            {
                Id = s.Id,
                UserId = s.UserId,
                PlanId = s.PlanId,
                PeriodEnd = s.PeriodEnd,
                AutoRenew = s.AutoRenew
            })
            .Take(batchSize)
            .ToListAsync(cancellationToken);
    }

    public Task UpdateRenewalBatchAsync(
        IReadOnlyCollection<SubscriptionRenewalUpdate> renewals,
        CancellationToken cancellationToken = default)
    {
        foreach (var renewal in renewals)
        {
            var subscription = new Subscription { Id = renewal.SubscriptionId };

            _context.Subscriptions.Attach(subscription);

            subscription.PeriodStart = renewal.NewPeriodStart;
            subscription.PeriodEnd = renewal.NewPeriodEnd;
            subscription.UpdatedAt = renewal.UpdatedAtUtc;

            _context.Entry(subscription).Property(s => s.PeriodStart).IsModified = true;
            _context.Entry(subscription).Property(s => s.PeriodEnd).IsModified = true;
            _context.Entry(subscription).Property(s => s.UpdatedAt).IsModified = true;
        }

        return Task.CompletedTask;
    }

    public async Task<int> ExpireSubscriptionsAsync(
        IReadOnlyCollection<Guid> subscriptionIds,
        DateTime updatedAtUtc,
        CancellationToken cancellationToken = default)
    {
        if (subscriptionIds.Count == 0)
        {
            return 0;
        }

        return await _context.Subscriptions
            .Where(s => subscriptionIds.Contains(s.Id))
            .ExecuteUpdateAsync(
                setters => setters
                    .SetProperty(s => s.Status, SubscriptionStatus.Expired)
                    .SetProperty(s => s.UpdatedAt, updatedAtUtc),
                cancellationToken);
    }
}
