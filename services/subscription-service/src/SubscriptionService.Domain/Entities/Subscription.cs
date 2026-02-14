using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Domain.Entities;

public class Subscription
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public Guid UserId { get; set; }
    public Guid PlanId { get; set; }
    public SubscriptionStatus Status { get; set; } = SubscriptionStatus.PendingPayment;
    public DateTime PeriodStart { get; set; }
    public DateTime PeriodEnd { get; set; }
    public bool AutoRenew { get; set; } = true;
    public DateTime? CancelledAt { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public Plan Plan { get; set; } = null!;
    public List<Payment> Payments { get; set; } = new();
    public List<Invoice> Invoices { get; set; } = new();
}
