using System.Text.Json;
using MediatR;
using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Events;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.Application.Commands.CancelSubscription;

public class CancelSubscriptionHandler
    : IRequestHandler<CancelSubscriptionCommand, CancelSubscriptionResult>
{
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IOutboxRepository _outboxRepository;
    private readonly ISubscriptionUnitOfWork _unitOfWork;
    private readonly ILogger<CancelSubscriptionHandler> _logger;

    public CancelSubscriptionHandler(
        ISubscriptionRepository subscriptionRepository,
        IOutboxRepository outboxRepository,
        ISubscriptionUnitOfWork unitOfWork,
        ILogger<CancelSubscriptionHandler> logger)
    {
        _subscriptionRepository = subscriptionRepository;
        _outboxRepository = outboxRepository;
        _unitOfWork = unitOfWork;
        _logger = logger;
    }

    public async Task<CancelSubscriptionResult> Handle(
        CancelSubscriptionCommand request,
        CancellationToken cancellationToken)
    {
        var subscription = await _subscriptionRepository.GetActiveByUserIdAsync(
            request.UserId, cancellationToken);

        if (subscription is null)
            throw new SubscriptionNotFoundException();

        if (subscription.Status != SubscriptionStatus.Active)
            throw new InvalidOperationException("Only active subscriptions can be cancelled.");

        var now = DateTime.UtcNow;
        subscription.Status = SubscriptionStatus.Cancelled;
        subscription.CancelledAt = now;
        subscription.AutoRenew = false;

        await _subscriptionRepository.UpdateAsync(subscription, cancellationToken);

        var cancelledEvent = new SubscriptionCancelledEvent(
            subscription.Id,
            subscription.UserId,
            now,
            subscription.PeriodEnd);

        await _outboxRepository.AddAsync(new OutboxMessage
        {
            EventType = "subscription.cancelled",
            Payload = JsonSerializer.Serialize(cancelledEvent),
            CreatedAt = now
        }, cancellationToken);

        await _unitOfWork.SaveChangesAsync(cancellationToken);

        _logger.LogInformation(
            "Subscription {SubscriptionId} cancelled for user {UserId}. Active until {PeriodEnd}",
            subscription.Id, subscription.UserId, subscription.PeriodEnd);

        return new CancelSubscriptionResult(
            subscription.Id,
            subscription.Status.ToString(),
            now,
            subscription.PeriodEnd);
    }
}
