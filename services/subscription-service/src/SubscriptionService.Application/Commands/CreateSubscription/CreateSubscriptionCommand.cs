using MediatR;

namespace SubscriptionService.Application.Commands.CreateSubscription;

public record CreateSubscriptionCommand(
    Guid UserId,
    Guid PlanId,
    string CardNumber
) : IRequest<CreateSubscriptionResult>;

public record CreateSubscriptionResult(
    Guid SubscriptionId,
    Guid UserId,
    Guid PlanId,
    string PlanName,
    string Tier,
    string Status,
    DateTime PeriodStart,
    DateTime PeriodEnd,
    decimal AmountCharged,
    string Currency
);
