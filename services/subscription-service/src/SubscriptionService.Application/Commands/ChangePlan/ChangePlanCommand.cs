using MediatR;

namespace SubscriptionService.Application.Commands.ChangePlan;

public record ChangePlanCommand(
    Guid UserId,
    Guid NewPlanId,
    string CardNumber
) : IRequest<ChangePlanResult>;

public record ChangePlanResult(
    Guid SubscriptionId,
    Guid OldPlanId,
    string OldPlanName,
    Guid NewPlanId,
    string NewPlanName,
    string NewTier,
    decimal PriceDifference,
    string Status
);
