namespace SubscriptionService.Domain.Events;

public record SubscriptionCreatedEvent(
    Guid SubscriptionId,
    Guid UserId,
    Guid PlanId,
    string PlanName,
    string Tier,
    DateTime PeriodStart,
    DateTime PeriodEnd
);
