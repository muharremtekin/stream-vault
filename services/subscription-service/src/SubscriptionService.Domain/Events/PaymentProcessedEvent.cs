namespace SubscriptionService.Domain.Events;

public record PaymentProcessedEvent(
    Guid PaymentId,
    Guid SubscriptionId,
    Guid UserId,
    decimal Amount,
    string Currency,
    string Status,
    DateTime ProcessedAt
);
