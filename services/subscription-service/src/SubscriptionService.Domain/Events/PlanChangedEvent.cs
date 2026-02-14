namespace SubscriptionService.Domain.Events;

public record PlanChangedEvent(
    Guid SubscriptionId,
    Guid UserId,
    Guid OldPlanId,
    string OldTier,
    Guid NewPlanId,
    string NewTier,
    DateTime ChangedAt
);
