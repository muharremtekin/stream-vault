namespace SubscriptionService.Domain.Events;

public record SubscriptionCancelledEvent(
    Guid SubscriptionId,
    Guid UserId,
    DateTime CancelledAt,
    DateTime PeriodEnd
);
