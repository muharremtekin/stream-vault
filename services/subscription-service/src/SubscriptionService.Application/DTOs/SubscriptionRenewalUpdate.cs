namespace SubscriptionService.Application.DTOs;

public class SubscriptionRenewalUpdate
{
    public Guid SubscriptionId { get; set; }
    public DateTime NewPeriodStart { get; set; }
    public DateTime NewPeriodEnd { get; set; }
    public DateTime UpdatedAtUtc { get; set; }
}
