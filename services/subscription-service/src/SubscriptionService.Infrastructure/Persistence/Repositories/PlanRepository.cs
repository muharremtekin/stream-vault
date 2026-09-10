using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class PlanRepository : IPlanRepository
{
    private readonly SubscriptionDbContext _context;

    public PlanRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task<Plan?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Plans
            .FirstOrDefaultAsync(p => p.Id == id, cancellationToken);
    }

    public async Task<List<Plan>> GetAllActiveAsync(CancellationToken cancellationToken = default)
    {
        return await _context.Plans
            .AsNoTracking()
            .Where(p => p.IsActive)
            .OrderBy(p => p.PriceMonthly)
            .ToListAsync(cancellationToken);
    }

    public async Task<List<PlanDto>> GetAllActiveDtosAsync(CancellationToken cancellationToken = default)
    {
        return await _context.Plans
            .AsNoTracking()
            .Where(p => p.IsActive)
            .OrderBy(p => p.PriceMonthly)
            .Select(p => new PlanDto
            {
                Id = p.Id,
                Name = p.Name,
                Tier = p.Tier.ToString(),
                PriceMonthly = p.PriceMonthly,
                MaxScreens = p.MaxScreens,
                MaxQuality = p.MaxQuality,
                Features = p.Features,
                IsActive = p.IsActive
            })
            .ToListAsync(cancellationToken);
    }

    public async Task<Plan?> GetByTierAsync(PlanTier tier, CancellationToken cancellationToken = default)
    {
        return await _context.Plans
            .FirstOrDefaultAsync(p => p.Tier == tier && p.IsActive, cancellationToken);
    }
}
