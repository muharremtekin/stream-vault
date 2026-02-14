using System.Text.Json;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Events;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.Application.Sagas;

public class ChangePlanSaga
{
    private readonly IPlanRepository _planRepository;
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IPaymentRepository _paymentRepository;
    private readonly IPaymentGateway _paymentGateway;
    private readonly IOutboxRepository _outboxRepository;
    private readonly ISagaRepository _sagaRepository;
    private readonly CompensatingActions _compensatingActions;
    private readonly ILogger<ChangePlanSaga> _logger;

    public ChangePlanSaga(
        IPlanRepository planRepository,
        ISubscriptionRepository subscriptionRepository,
        IPaymentRepository paymentRepository,
        IPaymentGateway paymentGateway,
        IOutboxRepository outboxRepository,
        ISagaRepository sagaRepository,
        CompensatingActions compensatingActions,
        ILogger<ChangePlanSaga> logger)
    {
        _planRepository = planRepository;
        _subscriptionRepository = subscriptionRepository;
        _paymentRepository = paymentRepository;
        _paymentGateway = paymentGateway;
        _outboxRepository = outboxRepository;
        _sagaRepository = sagaRepository;
        _compensatingActions = compensatingActions;
        _logger = logger;
    }

    public async Task<Subscription> ExecuteAsync(
        Guid subscriptionId,
        Guid newPlanId,
        string cardNumber,
        CancellationToken cancellationToken)
    {
        _logger.LogInformation(
            "Starting ChangePlan saga for subscription {SubscriptionId} to plan {NewPlanId}",
            subscriptionId, newPlanId);

        var (subscription, oldPlan, newPlan) = await ValidateChangeAsync(
            subscriptionId, newPlanId, cancellationToken);

        var isUpgrade = newPlan.PriceMonthly > oldPlan.PriceMonthly;
        var remainingDays = (decimal)(subscription.PeriodEnd - DateTime.UtcNow).TotalDays;
        var priceDifference = (newPlan.PriceMonthly - oldPlan.PriceMonthly) / 30m * remainingDays;
        priceDifference = Math.Round(priceDifference, 2);

        var sagaData = new ChangePlanSagaData
        {
            UserId = subscription.UserId,
            SubscriptionId = subscriptionId,
            OldPlanId = oldPlan.Id,
            NewPlanId = newPlanId,
            CardNumber = cardNumber,
            PriceDifference = priceDifference,
            IsUpgrade = isUpgrade
        };

        var saga = new SagaState
        {
            SagaType = "ChangePlan",
            CurrentStep = SagaStep.ValidateChange,
            Status = SagaStatus.InProgress,
            StateData = JsonSerializer.Serialize(sagaData)
        };
        await _sagaRepository.AddAsync(saga, cancellationToken);

        try
        {
            await ProcessPriceDifferenceAsync(
                saga, sagaData, subscription, cancellationToken);

            await UpdateSubscriptionAsync(
                saga, sagaData, subscription, newPlanId, cancellationToken);

            await PublishChangeEventsAsync(
                saga, sagaData, subscription, oldPlan, newPlan, cancellationToken);

            saga.Status = SagaStatus.Completed;
            await _sagaRepository.UpdateAsync(saga, cancellationToken);

            _logger.LogInformation(
                "ChangePlan saga {SagaId} completed for subscription {SubscriptionId}",
                saga.Id, subscriptionId);

            return subscription;
        }
        catch (PaymentFailedException)
        {
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex,
                "ChangePlan saga {SagaId} failed at step {Step}",
                saga.Id, saga.CurrentStep);

            saga.ErrorMessage = ex.Message;
            await _compensatingActions.CompensateChangePlanAsync(
                saga, sagaData, cancellationToken);

            throw new PaymentFailedException(
                $"Plan change failed at {saga.CurrentStep}: {ex.Message}");
        }
    }

    private async Task<(Subscription subscription, Plan oldPlan, Plan newPlan)> ValidateChangeAsync(
        Guid subscriptionId,
        Guid newPlanId,
        CancellationToken cancellationToken)
    {
        var subscription = await _subscriptionRepository.GetByIdAsync(
            subscriptionId, cancellationToken);

        if (subscription is null)
            throw new SubscriptionNotFoundException(subscriptionId);

        if (subscription.Status != SubscriptionStatus.Active)
            throw new InvalidOperationException("Only active subscriptions can change plans.");

        if (subscription.PlanId == newPlanId)
            throw new InvalidOperationException("Subscription is already on this plan.");

        var newPlan = await _planRepository.GetByIdAsync(newPlanId, cancellationToken);

        if (newPlan is null || !newPlan.IsActive)
            throw new PlanNotFoundException(newPlanId);

        return (subscription, subscription.Plan, newPlan);
    }

    private async Task ProcessPriceDifferenceAsync(
        SagaState saga,
        ChangePlanSagaData sagaData,
        Subscription subscription,
        CancellationToken cancellationToken)
    {
        if (sagaData.PaymentId.HasValue)
            return;

        saga.CurrentStep = SagaStep.ProcessPriceDifference;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        if (sagaData.IsUpgrade && sagaData.PriceDifference > 0)
        {
            var result = await _paymentGateway.ChargeAsync(
                sagaData.CardNumber,
                sagaData.PriceDifference,
                "TRY",
                cancellationToken);

            var payment = new Payment
            {
                SubscriptionId = subscription.Id,
                Amount = sagaData.PriceDifference,
                Currency = "TRY",
                Status = result.IsSuccess ? PaymentStatus.Succeeded : PaymentStatus.Failed,
                TransactionId = result.TransactionId,
                FailureReason = result.FailureReason,
                CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
            };

            await _paymentRepository.AddAsync(payment, cancellationToken);

            sagaData.PaymentId = payment.Id;
            sagaData.TransactionId = result.TransactionId;
            saga.StateData = JsonSerializer.Serialize(sagaData);
            await _sagaRepository.UpdateAsync(saga, cancellationToken);

            if (!result.IsSuccess)
            {
                _logger.LogWarning(
                    "Saga {SagaId}: Upgrade payment failed — {Reason}",
                    saga.Id, result.FailureReason);

                await _compensatingActions.CompensateChangePlanAsync(
                    saga, sagaData, cancellationToken);

                throw new PaymentFailedException(result.FailureReason ?? "Payment declined");
            }

            _logger.LogInformation(
                "Saga {SagaId}: Upgrade payment of {Amount} TRY succeeded",
                saga.Id, sagaData.PriceDifference);
        }
        else
        {
            var creditAmount = Math.Abs(sagaData.PriceDifference);
            var payment = new Payment
            {
                SubscriptionId = subscription.Id,
                Amount = -creditAmount,
                Currency = "TRY",
                Status = PaymentStatus.Succeeded,
                TransactionId = $"credit_{Guid.NewGuid():N}",
                CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
            };

            await _paymentRepository.AddAsync(payment, cancellationToken);

            sagaData.PaymentId = payment.Id;
            saga.StateData = JsonSerializer.Serialize(sagaData);
            await _sagaRepository.UpdateAsync(saga, cancellationToken);

            _logger.LogInformation(
                "Saga {SagaId}: Downgrade credit of {Amount} TRY recorded",
                saga.Id, creditAmount);
        }
    }

    private async Task UpdateSubscriptionAsync(
        SagaState saga,
        ChangePlanSagaData sagaData,
        Subscription subscription,
        Guid newPlanId,
        CancellationToken cancellationToken)
    {
        if (subscription.PlanId == newPlanId)
            return;

        saga.CurrentStep = SagaStep.UpdateSubscription;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        subscription.PlanId = newPlanId;
        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        var newPlan = await _planRepository.GetByIdAsync(newPlanId, cancellationToken);
        subscription.Plan = newPlan!;

        _logger.LogInformation(
            "Saga {SagaId}: Updated subscription {SubscriptionId} to plan {NewPlanId}",
            saga.Id, subscription.Id, newPlanId);
    }

    private async Task PublishChangeEventsAsync(
        SagaState saga,
        ChangePlanSagaData sagaData,
        Subscription subscription,
        Plan oldPlan,
        Plan newPlan,
        CancellationToken cancellationToken)
    {
        saga.CurrentStep = SagaStep.PublishChangeEvents;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        var now = DateTime.UtcNow;

        var planChangedEvent = new PlanChangedEvent(
            subscription.Id,
            subscription.UserId,
            oldPlan.Id,
            oldPlan.Tier.ToString(),
            newPlan.Id,
            newPlan.Tier.ToString(),
            now);

        await _outboxRepository.AddAsync(new OutboxMessage
        {
            EventType = "plan.changed",
            Payload = JsonSerializer.Serialize(planChangedEvent),
            CreatedAt = now
        }, cancellationToken);

        if (sagaData.PaymentId.HasValue)
        {
            var paymentEvent = new PaymentProcessedEvent(
                sagaData.PaymentId.Value,
                subscription.Id,
                subscription.UserId,
                sagaData.PriceDifference,
                "TRY",
                PaymentStatus.Succeeded.ToString(),
                now);

            await _outboxRepository.AddAsync(new OutboxMessage
            {
                EventType = "subscription.payment.processed",
                Payload = JsonSerializer.Serialize(paymentEvent),
                CreatedAt = now
            }, cancellationToken);
        }

        _logger.LogInformation(
            "Saga {SagaId}: Published outbox events for plan change {OldTier} → {NewTier}",
            saga.Id, oldPlan.Tier, newPlan.Tier);
    }
}
