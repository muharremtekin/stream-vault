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
        await _sagaRepository.AddAsync(saga, cancellationToken);

        try
        {
            var subscription = await CreatePendingSubscriptionAsync(
                saga, sagaData, plan, cancellationToken);

            await ProcessPaymentAsync(
                saga, sagaData, plan, subscription, cancellationToken);

            await ActivateSubscriptionAsync(
                saga, sagaData, subscription, cancellationToken);

            await CreateInvoiceAsync(
                saga, sagaData, plan, subscription, cancellationToken);

            await PublishEventsAsync(
                saga, sagaData, plan, subscription, cancellationToken);

            saga.Status = SagaStatus.Completed;
            await _sagaRepository.UpdateAsync(saga, cancellationToken);

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
            await _compensatingActions.CompensateCreateSubscriptionAsync(
                saga, sagaData, cancellationToken);

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

    private async Task<Subscription> CreatePendingSubscriptionAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        CancellationToken cancellationToken)
    {
        if (sagaData.SubscriptionId.HasValue)
        {
            var existing = await _subscriptionRepository.GetByIdAsync(
                sagaData.SubscriptionId.Value, cancellationToken);
            if (existing is not null) return existing;
        }

        saga.CurrentStep = SagaStep.CreateSubscription;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

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

        await _subscriptionRepository.AddAsync(subscription, cancellationToken);

        sagaData.SubscriptionId = subscription.Id;
        saga.StateData = JsonSerializer.Serialize(sagaData);
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Created pending subscription {SubscriptionId}",
            saga.Id, subscription.Id);

        return subscription;
    }

    private async Task ProcessPaymentAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        Subscription subscription,
        CancellationToken cancellationToken)
    {
        if (sagaData.PaymentId.HasValue)
            return;

        saga.CurrentStep = SagaStep.ProcessPayment;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        var result = await _paymentGateway.ChargeAsync(
            sagaData.CardNumber, plan.PriceMonthly, "TRY", cancellationToken);

        var payment = new Payment
        {
            SubscriptionId = subscription.Id,
            Amount = plan.PriceMonthly,
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
                "Saga {SagaId}: Payment failed — {Reason}",
                saga.Id, result.FailureReason);

            await _compensatingActions.CompensateCreateSubscriptionAsync(
                saga, sagaData, cancellationToken);

            throw new PaymentFailedException(result.FailureReason ?? "Payment declined");
        }

        _logger.LogInformation(
            "Saga {SagaId}: Payment {PaymentId} succeeded (txn: {TransactionId})",
            saga.Id, payment.Id, result.TransactionId);
    }

    private async Task ActivateSubscriptionAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Subscription subscription,
        CancellationToken cancellationToken)
    {
        if (subscription.Status == SubscriptionStatus.Active)
            return;

        saga.CurrentStep = SagaStep.Activate;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        var otherActive = await _subscriptionRepository.GetActiveByUserIdAsync(
            sagaData.UserId, cancellationToken);

        if (otherActive is not null && otherActive.Id != subscription.Id)
        {
            throw new ActiveSubscriptionExistsException(sagaData.UserId);
        }

        subscription.Status = SubscriptionStatus.Active;
        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Activated subscription {SubscriptionId}",
            saga.Id, subscription.Id);
    }

    private async Task CreateInvoiceAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        Subscription subscription,
        CancellationToken cancellationToken)
    {
        if (sagaData.InvoiceId.HasValue)
            return;

        saga.CurrentStep = SagaStep.CreateInvoice;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

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
        saga.StateData = JsonSerializer.Serialize(sagaData);
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        _logger.LogInformation(
            "Saga {SagaId}: Created invoice {InvoiceNumber} for subscription {SubscriptionId}",
            saga.Id, invoiceNumber, subscription.Id);
    }

    private async Task PublishEventsAsync(
        SagaState saga,
        CreateSubscriptionSagaData sagaData,
        Plan plan,
        Subscription subscription,
        CancellationToken cancellationToken)
    {
        saga.CurrentStep = SagaStep.PublishEvents;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

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

        if (sagaData.PaymentId.HasValue)
        {
            var paymentEvent = new PaymentProcessedEvent(
                sagaData.PaymentId.Value,
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
        }

        _logger.LogInformation(
            "Saga {SagaId}: Published outbox events for subscription {SubscriptionId}",
            saga.Id, subscription.Id);
    }
}
