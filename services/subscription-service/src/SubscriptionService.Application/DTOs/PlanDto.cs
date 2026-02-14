namespace SubscriptionService.Application.DTOs;

public class PlanDto
{
    public Guid Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Tier { get; set; } = string.Empty;
    public decimal PriceMonthly { get; set; }
    public int MaxScreens { get; set; }
    public string MaxQuality { get; set; } = string.Empty;
    public string Features { get; set; } = string.Empty;
    public bool IsActive { get; set; }
}
