using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Infrastructure.Persistence;

public class DbSeeder
{
    private readonly SubscriptionDbContext _context;
    private readonly ILogger<DbSeeder> _logger;

    private static readonly Plan[] SeedPlans =
    [
        new()
        {
            Id = Guid.Parse("a1b2c3d4-0001-0001-0001-000000000001"),
            Name = "Basic",
            Tier = PlanTier.Basic,
            PriceMonthly = 49.99m,
            MaxScreens = 1,
            MaxQuality = "720p",
            Features = "[\"Ad-supported\",\"1 screen\",\"720p max quality\",\"Mobile & tablet\"]",
            IsActive = true
        },
        new()
        {
            Id = Guid.Parse("a1b2c3d4-0001-0001-0001-000000000002"),
            Name = "Standard",
            Tier = PlanTier.Standard,
            PriceMonthly = 79.99m,
            MaxScreens = 2,
            MaxQuality = "1080p",
            Features = "[\"Ad-free\",\"2 screens\",\"1080p max quality\",\"Mobile, tablet & computer\",\"Downloads\"]",
            IsActive = true
        },
        new()
        {
            Id = Guid.Parse("a1b2c3d4-0001-0001-0001-000000000003"),
            Name = "Premium",
            Tier = PlanTier.Premium,
            PriceMonthly = 119.99m,
            MaxScreens = 4,
            MaxQuality = "4K",
            Features = "[\"Ad-free\",\"4 screens\",\"4K+HDR max quality\",\"All devices\",\"Downloads\",\"Spatial Audio\"]",
            IsActive = true
        }
    ];

    public DbSeeder(SubscriptionDbContext context, ILogger<DbSeeder> logger)
    {
        _context = context;
        _logger = logger;
    }

    public async Task SeedAsync(CancellationToken cancellationToken = default)
    {
        foreach (var plan in SeedPlans)
        {
            var exists = await _context.Plans.AnyAsync(p => p.Tier == plan.Tier, cancellationToken);
            if (exists)
            {
                _logger.LogDebug("Seed plan {PlanName} already exists. Skipping.", plan.Name);
                continue;
            }

            _context.Plans.Add(plan);
            _logger.LogInformation("Seeded plan {PlanName} ({Tier}) at {Price} TRY/month.",
                plan.Name, plan.Tier, plan.PriceMonthly);
        }

        await _context.SaveChangesAsync(cancellationToken);
    }
}
