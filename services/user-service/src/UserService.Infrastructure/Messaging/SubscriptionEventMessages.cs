namespace UserService.Infrastructure.Messaging;

public record SubscriptionCreatedMessage(
    Guid SubscriptionId,
    Guid UserId,
    Guid PlanId,
    string PlanName,
    string Tier,
    DateTime PeriodStart,
    DateTime PeriodEnd
);

public record SubscriptionCancelledMessage(
    Guid SubscriptionId,
    Guid UserId,
    DateTime CancelledAt,
    DateTime PeriodEnd
);

public record PlanChangedMessage(
    Guid SubscriptionId,
    Guid UserId,
    Guid OldPlanId,
    string OldTier,
    Guid NewPlanId,
    string NewTier,
    DateTime ChangedAt
);
