using MediatR;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Sagas;

namespace SubscriptionService.Application.Commands.CreateSubscription;

public class CreateSubscriptionHandler
    : IRequestHandler<CreateSubscriptionCommand, CreateSubscriptionResult>
{
    private readonly SubscriptionSaga _saga;
    private readonly ILogger<CreateSubscriptionHandler> _logger;

    public CreateSubscriptionHandler(
        SubscriptionSaga saga,
        ILogger<CreateSubscriptionHandler> logger)
    {
        _saga = saga;
        _logger = logger;
    }

    public async Task<CreateSubscriptionResult> Handle(
        CreateSubscriptionCommand request,
        CancellationToken cancellationToken)
    {
        _logger.LogInformation(
            "Creating subscription for user {UserId} on plan {PlanId}",
            request.UserId, request.PlanId);

        var subscription = await _saga.ExecuteAsync(
            request.UserId,
            request.PlanId,
            request.CardNumber,
            cancellationToken);

        return new CreateSubscriptionResult(
            subscription.Id,
            subscription.UserId,
            subscription.PlanId,
            subscription.Plan.Name,
            subscription.Plan.Tier.ToString(),
            subscription.Status.ToString(),
            subscription.PeriodStart,
            subscription.PeriodEnd,
            subscription.Plan.PriceMonthly,
            "TRY");
    }
}
