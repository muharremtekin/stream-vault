using System.Text.Json;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Events;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.Application.Sagas;

public class SubscriptionSaga
{
    private readonly IPlanRepository _planRepository;
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IPaymentRepository _paymentRepository;
    private readonly IPaymentGateway _paymentGateway;
    private readonly IInvoiceRepository _invoiceRepository;
    private readonly IOutboxRepository _outboxRepository;
    private readonly ISagaRepository _sagaRepository;
    private readonly ISubscriptionUnitOfWork _unitOfWork;
    private readonly CompensatingActions _compensatingActions;
    private readonly ILogger<SubscriptionSaga> _logger;

    public SubscriptionSaga(
        IPlanRepository planRepository,
        ISubscriptionRepository subscriptionRepository,
        IPaymentRepository paymentRepository,
        IPaymentGateway paymentGateway,
        IInvoiceRepository invoiceRepository,
        IOutboxRepository outboxRepository,
        ISagaRepository sagaRepository,
        ISubscriptionUnitOfWork unitOfWork,
        CompensatingActions compensatingActions,
        ILogger<SubscriptionSaga> logger)
    {
        _planRepository = planRepository;
        _subscriptionRepository = subscriptionRepository;
        _paymentRepository = paymentRepository;
        _paymentGateway = paymentGateway;
        _invoiceRepository = invoiceRepository;
        _outboxRepository = outboxRepository;
        _sagaRepository = sagaRepository;
        _unitOfWork = unitOfWork;
        _compensatingActions = compensatingActions;
        _logger = logger;
    }

    public async Task<Subscription> ExecuteAsync(
        Guid userId,
        Guid planId,
        string cardNumber,
        CancellationToken cancellationToken)
    {
        _logger.LogInformation(
            "Starting CreateSubscription saga for user {UserId} on plan {PlanId}",
            userId, planId);

        var plan = await ValidatePlanAsync(userId, planId, cancellationToken);

        var sagaData = new CreateSubscriptionSagaData
        {
            UserId = userId,
            PlanId = planId,
            CardNumber = cardNumber
        };

        var saga = new SagaState
        {
            SagaType = "CreateSubscription",
            CurrentStep = SagaStep.ValidatePlan,
            Status = SagaStatus.InProgress,
            StateData = JsonSerializer.Serialize(sagaData)
        };

        var pendingPhaseCommitted = false;

        try
        {
            var subscription = await CreatePendingSubscriptionPhaseAsync(
                saga, sagaData, plan, cancellationToken);
            pendingPhaseCommitted = true;

            var paymentResult = await _paymentGateway.ChargeAsync(
                sagaData.CardNumber,
                plan.PriceMonthly,
                "TRY",
                cancellationToken);

            if (!paymentResult.IsSuccess)
            {
                await PersistFailedPaymentPhaseAsync(
                    saga,
                    sagaData,
                    subscription,
                    plan,
                    paymentResult,
                    cancellationToken);

                throw new PaymentFailedException(paymentResult.FailureReason ?? "Payment declined");
            }

            await FinalizeSuccessfulSubscriptionPhaseAsync(
                saga,
                sagaData,
                plan,
                subscription,
                paymentResult,
                cancellationToken);

            _logger.LogInformation(
                "CreateSubscription saga {SagaId} completed for user {UserId}",
                saga.Id, userId);

            return subscription;
        }
        catch (PaymentFailedException)
        {
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex,
                "CreateSubscription saga {SagaId} failed at step {Step}",
                saga.Id, saga.CurrentStep);

            saga.ErrorMessage = ex.Message;

            if (pendingPhaseCommitted)
            {
                await _compensatingActions.CompensateCreateSubscriptionAsync(
                    saga, sagaData, cancellationToken);
            }

            throw new PaymentFailedException(
                $"Subscription creation failed at {saga.CurrentStep}: {ex.Message}");
        }
    }

    private async Task<Plan> ValidatePlanAsync(
        Guid userId,
        Guid planId,
        CancellationToken cancellationToken)
    {
        var plan = await _planRepository.GetByIdAsync(planId, cancellationToken);

        if (plan is null || !plan.IsActive)
            throw new PlanNotFoundException(planId);

        var existingSub = await _subscriptionRepository.GetActiveByUserIdAsync(
            userId, cancellationToken);

        if (existingSub is not null)
            throw new ActiveSubscriptionExistsException(userId);

        return plan;
    }

    private async Task<Subscription> CreatePendingSubscriptionPhaseAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        CancellationToken cancellationToken)
    {
        var now = DateTime.UtcNow;
        var subscription = new Subscription
        {
            UserId = sagaData.UserId,
            PlanId = plan.Id,
            Status = SubscriptionStatus.PendingPayment,
            PeriodStart = now,
            PeriodEnd = now.AddMonths(1),
            AutoRenew = true,
            CreatedAt = now,
            UpdatedAt = now,
            Plan = plan
        };

        sagaData.SubscriptionId = subscription.Id;
        saga.CurrentStep = SagaStep.ProcessPayment;
        saga.StateData = JsonSerializer.Serialize(sagaData);

        await _sagaRepository.AddAsync(saga, cancellationToken);
        await _subscriptionRepository.AddAsync(subscription, cancellationToken);
        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Committed pre-payment phase for subscription {SubscriptionId}",
            saga.Id, subscription.Id);

        return subscription;
    }

    private async Task PersistFailedPaymentPhaseAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Subscription subscription,
        Plan plan,
        PaymentResult paymentResult,
        CancellationToken cancellationToken)
    {
        saga.CurrentStep = SagaStep.ProcessPayment;

        var payment = new Payment
        {
            SubscriptionId = subscription.Id,
            Amount = plan.PriceMonthly,
            Currency = "TRY",
            Status = PaymentStatus.Failed,
            TransactionId = paymentResult.TransactionId,
            FailureReason = paymentResult.FailureReason,
            CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
        };

        await _paymentRepository.AddAsync(payment, cancellationToken);

        sagaData.PaymentId = payment.Id;
        sagaData.TransactionId = paymentResult.TransactionId;

        subscription.Status = SubscriptionStatus.Failed;
        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        saga.Status = SagaStatus.Failed;
        saga.ErrorMessage = paymentResult.FailureReason;
        saga.StateData = JsonSerializer.Serialize(sagaData);
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogWarning(
            "Saga {SagaId}: Persisted failed payment for subscription {SubscriptionId}",
            saga.Id, subscription.Id);
    }

    private async Task FinalizeSuccessfulSubscriptionPhaseAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        Subscription subscription,
        PaymentResult paymentResult,
        CancellationToken cancellationToken)
    {
        var otherActive = await _subscriptionRepository.GetActiveByUserIdAsync(
            sagaData.UserId, cancellationToken);

        if (otherActive is not null && otherActive.Id != subscription.Id)
        {
            throw new ActiveSubscriptionExistsException(sagaData.UserId);
        }

        var payment = new Payment
        {
            SubscriptionId = subscription.Id,
            Amount = plan.PriceMonthly,
            Currency = "TRY",
            Status = PaymentStatus.Succeeded,
            TransactionId = paymentResult.TransactionId,
            FailureReason = paymentResult.FailureReason,
            CardLastFour = sagaData.CardNumber.Replace(" ", "").Replace("-", "")[^4..]
        };
        await _paymentRepository.AddAsync(payment, cancellationToken);
        sagaData.PaymentId = payment.Id;
        sagaData.TransactionId = paymentResult.TransactionId;

        saga.CurrentStep = SagaStep.Activate;
        subscription.Status = SubscriptionStatus.Active;
        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        saga.CurrentStep = SagaStep.CreateInvoice;

        var invoiceNumber = await _invoiceRepository.GenerateInvoiceNumberAsync(cancellationToken);

        var invoice = new Invoice
        {
            SubscriptionId = subscription.Id,
            InvoiceNumber = invoiceNumber,
            Amount = plan.PriceMonthly,
            Currency = "TRY",
            PeriodStart = subscription.PeriodStart,
            PeriodEnd = subscription.PeriodEnd,
            IssuedAt = DateTime.UtcNow
        };

        await _invoiceRepository.AddAsync(invoice, cancellationToken);
        sagaData.InvoiceId = invoice.Id;

        saga.CurrentStep = SagaStep.PublishEvents;

        var now = DateTime.UtcNow;

        var subscriptionEvent = new SubscriptionCreatedEvent(
            subscription.Id,
            subscription.UserId,
            plan.Id,
            plan.Name,
            plan.Tier.ToString(),
            subscription.PeriodStart,
            subscription.PeriodEnd);

        await _outboxRepository.AddAsync(new OutboxMessage
        {
            EventType = "subscription.created",
            Payload = JsonSerializer.Serialize(subscriptionEvent),
            CreatedAt = now
        }, cancellationToken);

        var paymentEvent = new PaymentProcessedEvent(
            payment.Id,
            subscription.Id,
            subscription.UserId,
            plan.PriceMonthly,
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
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Committed activation, invoice and outbox phase for subscription {SubscriptionId}",
            saga.Id, subscription.Id);
    }
}
