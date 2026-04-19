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
    private readonly ISubscriptionUnitOfWork _unitOfWork;
    private readonly CompensatingActions _compensatingActions;
    private readonly ILogger<ChangePlanSaga> _logger;

    public ChangePlanSaga(
        IPlanRepository planRepository,
        ISubscriptionRepository subscriptionRepository,
        IPaymentRepository paymentRepository,
        IPaymentGateway paymentGateway,
        IOutboxRepository outboxRepository,
        ISagaRepository sagaRepository,
        ISubscriptionUnitOfWork unitOfWork,
        CompensatingActions compensatingActions,
        ILogger<ChangePlanSaga> logger)
    {
        _planRepository = planRepository;
        _subscriptionRepository = subscriptionRepository;
        _paymentRepository = paymentRepository;
        _paymentGateway = paymentGateway;
        _outboxRepository = outboxRepository;
        _sagaRepository = sagaRepository;
        _unitOfWork = unitOfWork;
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

        var prePaymentPhaseCommitted = false;

        try
        {
            await _sagaRepository.AddAsync(saga, cancellationToken);

            PaymentResult? paymentResult = null;

            if (isUpgrade && priceDifference > 0)
            {
                saga.CurrentStep = SagaStep.ProcessPriceDifference;
                saga.StateData = JsonSerializer.Serialize(sagaData);
                await _unitOfWork.SaveChangesAsync(cancellationToken);
                prePaymentPhaseCommitted = true;

                paymentResult = await _paymentGateway.ChargeAsync(
                    sagaData.CardNumber,
                    sagaData.PriceDifference,
                    "TRY",
                    cancellationToken);

                if (!paymentResult.IsSuccess)
                {
                    await PersistFailedUpgradePaymentPhaseAsync(
                        saga, sagaData, subscription, paymentResult, cancellationToken);

                    throw new PaymentFailedException(paymentResult.FailureReason ?? "Payment declined");
                }
            }

            await FinalizePlanChangePhaseAsync(
                saga,
                sagaData,
                subscription,
                oldPlan,
                newPlan,
                paymentResult,
                prePaymentPhaseCommitted,
                cancellationToken);

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

            if (prePaymentPhaseCommitted)
            {
                await _compensatingActions.CompensateChangePlanAsync(
                    saga, sagaData, cancellationToken);
            }

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

    private async Task PersistFailedUpgradePaymentPhaseAsync(
        SagaState saga,
        ChangePlanSagaData sagaData,
        Subscription subscription,
        PaymentResult paymentResult,
        CancellationToken cancellationToken)
    {
        saga.CurrentStep = SagaStep.ProcessPriceDifference;

        var payment = new Payment
        {
            SubscriptionId = subscription.Id,
            Amount = sagaData.PriceDifference,
            Currency = "TRY",
            Status = PaymentStatus.Failed,
            TransactionId = paymentResult.TransactionId,
            FailureReason = paymentResult.FailureReason,
            CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
        };

        await _paymentRepository.AddAsync(payment, cancellationToken);

        sagaData.PaymentId = payment.Id;
        sagaData.TransactionId = paymentResult.TransactionId;
        saga.Status = SagaStatus.Failed;
        saga.ErrorMessage = paymentResult.FailureReason;
        saga.StateData = JsonSerializer.Serialize(sagaData);
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogWarning(
            "Saga {SagaId}: Persisted failed upgrade payment for subscription {SubscriptionId}. WriteOperations={WriteOperations}",
            saga.Id,
            subscription.Id,
            2);
    }

    private async Task FinalizePlanChangePhaseAsync(
        SagaState saga,
        ChangePlanSagaData sagaData,
        Subscription subscription,
        Plan oldPlan,
        Plan newPlan,
        PaymentResult? paymentResult,
        bool sagaAlreadyCommitted,
        CancellationToken cancellationToken)
    {
        if (paymentResult is not null)
        {
            var payment = new Payment
            {
                SubscriptionId = subscription.Id,
                Amount = sagaData.PriceDifference,
                Currency = "TRY",
                Status = PaymentStatus.Succeeded,
                TransactionId = paymentResult.TransactionId,
                FailureReason = paymentResult.FailureReason,
                CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
            };

            await _paymentRepository.AddAsync(payment, cancellationToken);
            sagaData.PaymentId = payment.Id;
            sagaData.TransactionId = paymentResult.TransactionId;
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
        }

        saga.CurrentStep = SagaStep.UpdateSubscription;
        subscription.PlanId = newPlan.Id;
        subscription.Plan = newPlan;
        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        saga.CurrentStep = SagaStep.PublishChangeEvents;

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

        var paymentEvent = new PaymentProcessedEvent(
            sagaData.PaymentId!.Value,
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

        saga.Status = SagaStatus.Completed;
        saga.StateData = JsonSerializer.Serialize(sagaData);

        if (sagaAlreadyCommitted)
        {
            await _sagaRepository.UpdateAsync(saga, cancellationToken);
        }

        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Committed subscription, payment and outbox phase for plan change {OldTier} to {NewTier}. WriteOperations={WriteOperations}",
            saga.Id,
            oldPlan.Tier,
            newPlan.Tier,
            5);
    }
}
