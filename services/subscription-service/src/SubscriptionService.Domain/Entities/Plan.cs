using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Domain.Entities;

public class Plan
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public string Name { get; set; } = string.Empty;
    public PlanTier Tier { get; set; }
    public decimal PriceMonthly { get; set; }
    public int MaxScreens { get; set; }
    public string MaxQuality { get; set; } = string.Empty;
    public string Features { get; set; } = string.Empty;
    public bool IsActive { get; set; } = true;
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public List<Subscription> Subscriptions { get; set; } = new();
}
