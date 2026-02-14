namespace SubscriptionService.Domain.Enums;

public enum SubscriptionStatus
{
    PendingPayment = 0,
    Active = 1,
    Cancelled = 2,
    Expired = 3,
    Failed = 4
}
