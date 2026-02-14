namespace SubscriptionService.Application.Sagas;

public class CreateSubscriptionSagaData
{
    public Guid UserId { get; set; }
    public Guid PlanId { get; set; }
    public string CardNumber { get; set; } = string.Empty;
    public Guid? SubscriptionId { get; set; }
    public Guid? PaymentId { get; set; }
    public Guid? InvoiceId { get; set; }
    public string? TransactionId { get; set; }
}

public class ChangePlanSagaData
{
    public Guid UserId { get; set; }
    public Guid SubscriptionId { get; set; }
    public Guid OldPlanId { get; set; }
    public Guid NewPlanId { get; set; }
    public string CardNumber { get; set; } = string.Empty;
    public decimal PriceDifference { get; set; }
    public bool IsUpgrade { get; set; }
    public Guid? PaymentId { get; set; }
    public string? TransactionId { get; set; }
}
