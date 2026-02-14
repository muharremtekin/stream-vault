using System.Text.Json;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Application.Sagas;

public class CompensatingActions
{
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IPaymentGateway _paymentGateway;
    private readonly ISagaRepository _sagaRepository;
    private readonly ILogger<CompensatingActions> _logger;

    public CompensatingActions(
        ISubscriptionRepository subscriptionRepository,
        IPaymentGateway paymentGateway,
        ISagaRepository sagaRepository,
        ILogger<CompensatingActions> logger)
    {
        _subscriptionRepository = subscriptionRepository;
        _paymentGateway = paymentGateway;
        _sagaRepository = sagaRepository;
        _logger = logger;
    }

    public async Task CompensateCreateSubscriptionAsync(
        SagaState saga,
        CreateSubscriptionSagaData data,
        CancellationToken cancellationToken)
    {
        _logger.LogWarning(
            "Compensating CreateSubscription saga {SagaId} at step {Step}",
            saga.Id, saga.CurrentStep);

        saga.Status = SagaStatus.Compensating;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        if (data.TransactionId is not null)
        {
            _logger.LogInformation(
                "Refunding transaction {TransactionId} for saga {SagaId}",
                data.TransactionId, saga.Id);

            await _paymentGateway.RefundAsync(
                data.TransactionId, 0, cancellationToken);
        }

        if (data.SubscriptionId.HasValue)
        {
            var subscription = await _subscriptionRepository.GetByIdAsync(
                data.SubscriptionId.Value, cancellationToken);

            if (subscription is not null && subscription.Status != SubscriptionStatus.Failed)
            {
                subscription.Status = SubscriptionStatus.Failed;
                await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

                _logger.LogInformation(
                    "Set subscription {SubscriptionId} to Failed for saga {SagaId}",
                    data.SubscriptionId, saga.Id);
            }
        }

        saga.Status = SagaStatus.Failed;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        _logger.LogWarning("CreateSubscription saga {SagaId} compensation completed", saga.Id);
    }

    public async Task CompensateChangePlanAsync(
        SagaState saga,
        ChangePlanSagaData data,
        CancellationToken cancellationToken)
    {
        _logger.LogWarning(
            "Compensating ChangePlan saga {SagaId} at step {Step}",
            saga.Id, saga.CurrentStep);

        saga.Status = SagaStatus.Compensating;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        if (data.TransactionId is not null)
        {
            _logger.LogInformation(
                "Refunding transaction {TransactionId} for saga {SagaId}",
                data.TransactionId, saga.Id);

            await _paymentGateway.RefundAsync(
                data.TransactionId, Math.Abs(data.PriceDifference), cancellationToken);
        }

        saga.Status = SagaStatus.Failed;
        await _sagaRepository.UpdateAsync(saga, cancellationToken);

        _logger.LogWarning("ChangePlan saga {SagaId} compensation completed", saga.Id);
    }
}
