using MediatR;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Application.Sagas;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.Application.Commands.ChangePlan;

public class ChangePlanHandler
    : IRequestHandler<ChangePlanCommand, ChangePlanResult>
{
    private readonly ChangePlanSaga _saga;
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IPlanRepository _planRepository;
    private readonly ILogger<ChangePlanHandler> _logger;

    public ChangePlanHandler(
        ChangePlanSaga saga,
        ISubscriptionRepository subscriptionRepository,
        IPlanRepository planRepository,
        ILogger<ChangePlanHandler> logger)
    {
        _saga = saga;
        _subscriptionRepository = subscriptionRepository;
        _planRepository = planRepository;
        _logger = logger;
    }

    public async Task<ChangePlanResult> Handle(
        ChangePlanCommand request,
        CancellationToken cancellationToken)
    {
        var subscription = await _subscriptionRepository.GetActiveByUserIdAsync(
            request.UserId, cancellationToken);

        if (subscription is null)
            throw new SubscriptionNotFoundException();

        var oldPlan = subscription.Plan;

        _logger.LogInformation(
            "Changing plan for subscription {SubscriptionId} from {OldPlan} to {NewPlan}",
            subscription.Id, oldPlan.Name, request.NewPlanId);

        var updatedSubscription = await _saga.ExecuteAsync(
            subscription.Id,
            request.NewPlanId,
            request.CardNumber,
            cancellationToken);

        var newPlan = updatedSubscription.Plan;

        var remainingDays = (decimal)(subscription.PeriodEnd - DateTime.UtcNow).TotalDays;
        var priceDifference = (newPlan.PriceMonthly - oldPlan.PriceMonthly) / 30m * remainingDays;

        return new ChangePlanResult(
            updatedSubscription.Id,
            oldPlan.Id,
            oldPlan.Name,
            newPlan.Id,
            newPlan.Name,
            newPlan.Tier.ToString(),
            Math.Round(priceDifference, 2),
            updatedSubscription.Status.ToString());
    }
}
