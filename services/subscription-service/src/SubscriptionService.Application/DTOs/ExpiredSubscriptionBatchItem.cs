namespace SubscriptionService.Application.DTOs;

public class ExpiredSubscriptionBatchItem
{
    public Guid Id { get; set; }
    public Guid UserId { get; set; }
    public Guid PlanId { get; set; }
    public DateTime PeriodEnd { get; set; }
    public bool AutoRenew { get; set; }
}
