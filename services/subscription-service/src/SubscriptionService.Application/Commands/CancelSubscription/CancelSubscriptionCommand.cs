using MediatR;

namespace SubscriptionService.Application.Commands.CancelSubscription;

public record CancelSubscriptionCommand(
    Guid UserId
) : IRequest<CancelSubscriptionResult>;

public record CancelSubscriptionResult(
    Guid SubscriptionId,
    string Status,
    DateTime CancelledAt,
    DateTime PeriodEnd
);
