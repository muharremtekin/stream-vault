using System.Text.Json;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Api.BackgroundServices;

public class SubscriptionRenewalService : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly ILogger<SubscriptionRenewalService> _logger;
    private readonly TimeSpan _interval = TimeSpan.FromHours(1);

    public SubscriptionRenewalService(
        IServiceProvider serviceProvider,
        ILogger<SubscriptionRenewalService> logger)
    {
        _serviceProvider = serviceProvider;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Subscription renewal service started.");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessExpiredSubscriptionsAsync(stoppingToken);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error processing expired subscriptions.");
            }

            await Task.Delay(_interval, stoppingToken);
        }

        _logger.LogInformation("Subscription renewal service stopped.");
    }

    private async Task ProcessExpiredSubscriptionsAsync(CancellationToken cancellationToken)
    {
        using var scope = _serviceProvider.CreateScope();
        var subscriptionRepository = scope.ServiceProvider.GetRequiredService<ISubscriptionRepository>();
        var outboxRepository = scope.ServiceProvider.GetRequiredService<IOutboxRepository>();
        var unitOfWork = scope.ServiceProvider.GetRequiredService<ISubscriptionUnitOfWork>();

        var expiredSubscriptions = await subscriptionRepository.GetExpiredSubscriptionsAsync(cancellationToken);
        if (expiredSubscriptions.Count == 0)
            return;

        _logger.LogInformation("Found {Count} expired subscriptions to process.", expiredSubscriptions.Count);

        foreach (var subscription in expiredSubscriptions)
        {
            try
            {
                if (subscription.AutoRenew)
                {
                    subscription.PeriodStart = subscription.PeriodEnd;
                    subscription.PeriodEnd = subscription.PeriodEnd.AddDays(30);
                    subscription.UpdatedAt = DateTime.UtcNow;

                    await subscriptionRepository.UpdateAsync(subscription, cancellationToken);

                    var payload = JsonSerializer.Serialize(new
                    {
                        eventId = Guid.NewGuid(),
                        eventType = "subscription.renewed",
                        timestamp = DateTime.UtcNow,
                        source = "subscription-service",
                        data = new
                        {
                            subscriptionId = subscription.Id,
                            userId = subscription.UserId,
                            planId = subscription.PlanId,
                            periodStart = subscription.PeriodStart,
                            periodEnd = subscription.PeriodEnd
                        }
                    });

                    await outboxRepository.AddAsync(new Domain.Entities.OutboxMessage
                    {
                        EventType = "subscription.renewed",
                        Payload = payload,
                    }, cancellationToken);

                    await unitOfWork.SaveChangesAsync(cancellationToken);

                    _logger.LogInformation(
                        "Renewed subscription {SubscriptionId} for user {UserId}. New period: {Start} - {End}.",
                        subscription.Id, subscription.UserId, subscription.PeriodStart, subscription.PeriodEnd);
                }
                else
                {
                    subscription.Status = SubscriptionStatus.Expired;
                    subscription.UpdatedAt = DateTime.UtcNow;

                    await subscriptionRepository.UpdateAsync(subscription, cancellationToken);
                    await unitOfWork.SaveChangesAsync(cancellationToken);

                    _logger.LogInformation(
                        "Expired subscription {SubscriptionId} for user {UserId} (auto-renew disabled).",
                        subscription.Id, subscription.UserId);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex,
                    "Failed to process expired subscription {SubscriptionId}.",
                    subscription.Id);
            }
        }
    }
}
