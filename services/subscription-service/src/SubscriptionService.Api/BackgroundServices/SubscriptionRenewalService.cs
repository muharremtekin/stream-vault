using System.Diagnostics;
using System.Diagnostics.Metrics;
using System.Text.Json;
using Microsoft.Extensions.Options;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Api.BackgroundServices;

public class SubscriptionRenewalService : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly ILogger<SubscriptionRenewalService> _logger;
    private readonly SubscriptionRenewalOptions _options;

    private static readonly Meter Meter = new("SubscriptionService.Renewal");
    private static readonly Counter<long> BatchCounter = Meter.CreateCounter<long>("subscription.renewal.batches");
    private static readonly Counter<long> ProcessedCounter = Meter.CreateCounter<long>("subscription.renewal.processed");
    private static readonly Histogram<double> BatchDurationMs = Meter.CreateHistogram<double>("subscription.renewal.batch.duration.ms");

    public SubscriptionRenewalService(
        IServiceProvider serviceProvider,
        ILogger<SubscriptionRenewalService> logger,
        IOptions<SubscriptionRenewalOptions> options)
    {
        _serviceProvider = serviceProvider;
        _logger = logger;
        _options = options.Value;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation(
            "Subscription renewal service started. Interval={IntervalSeconds}s, BatchSize={BatchSize}",
            _options.IntervalSeconds,
            _options.BatchSize);

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

            await Task.Delay(TimeSpan.FromSeconds(_options.IntervalSeconds), stoppingToken);
        }

        _logger.LogInformation("Subscription renewal service stopped.");
    }

    private async Task ProcessExpiredSubscriptionsAsync(CancellationToken cancellationToken)
    {
        var totalProcessed = 0;
        var totalRenewed = 0;
        var totalExpired = 0;
        var batchNumber = 0;

        while (!cancellationToken.IsCancellationRequested)
        {
            var stopwatch = Stopwatch.StartNew();
            using var scope = _serviceProvider.CreateScope();
            var subscriptionRepository = scope.ServiceProvider.GetRequiredService<ISubscriptionRepository>();
            var outboxRepository = scope.ServiceProvider.GetRequiredService<IOutboxRepository>();
            var unitOfWork = scope.ServiceProvider.GetRequiredService<ISubscriptionUnitOfWork>();

            var expiredSubscriptions = await subscriptionRepository
                .GetExpiredSubscriptionsAsync(_options.BatchSize, cancellationToken);

            if (expiredSubscriptions.Count == 0)
            {
                break;
            }

            batchNumber++;
            totalProcessed += expiredSubscriptions.Count;

            var processedAtUtc = DateTime.UtcNow;
            var toExpire = expiredSubscriptions
                .Where(subscription => !subscription.AutoRenew)
                .Select(subscription => subscription.Id)
                .ToArray();
            var renewals = expiredSubscriptions
                .Where(subscription => subscription.AutoRenew)
                .Select(subscription => new SubscriptionRenewalUpdate
                {
                    SubscriptionId = subscription.Id,
                    NewPeriodStart = subscription.PeriodEnd,
                    NewPeriodEnd = subscription.PeriodEnd.AddDays(30),
                    UpdatedAtUtc = processedAtUtc
                })
                .ToArray();

            if (toExpire.Length > 0)
            {
                var expiredCount = await subscriptionRepository
                    .ExpireSubscriptionsAsync(toExpire, processedAtUtc, cancellationToken);
                totalExpired += expiredCount;
            }

            if (renewals.Length > 0)
            {
                await subscriptionRepository.UpdateRenewalBatchAsync(renewals, cancellationToken);

                foreach (var renewal in renewals)
                {
                    var source = expiredSubscriptions.Single(subscription => subscription.Id == renewal.SubscriptionId);
                    await outboxRepository.AddAsync(new Domain.Entities.OutboxMessage
                    {
                        EventType = "subscription.renewed",
                        Payload = BuildRenewalPayload(source, renewal.NewPeriodStart, renewal.NewPeriodEnd)
                    }, cancellationToken);
                }

                await unitOfWork.SaveChangesAsync(cancellationToken);
                totalRenewed += renewals.Length;
            }

            stopwatch.Stop();
            BatchCounter.Add(1);
            ProcessedCounter.Add(expiredSubscriptions.Count);
            BatchDurationMs.Record(stopwatch.Elapsed.TotalMilliseconds);

            _logger.LogInformation(
                "Processed renewal batch {BatchNumber}. Retrieved={RetrievedCount}, Renewed={RenewedCount}, Expired={ExpiredCount}, DurationMs={DurationMs}.",
                batchNumber,
                expiredSubscriptions.Count,
                renewals.Length,
                toExpire.Length,
                stopwatch.Elapsed.TotalMilliseconds);

            if (expiredSubscriptions.Count < _options.BatchSize)
            {
                break;
            }
        }

        if (totalProcessed > 0)
        {
            _logger.LogInformation(
                "Completed subscription renewal cycle. Processed={ProcessedCount}, Renewed={RenewedCount}, Expired={ExpiredCount}, Batches={BatchCount}.",
                totalProcessed,
                totalRenewed,
                totalExpired,
                batchNumber);
        }
    }

    private static string BuildRenewalPayload(
        ExpiredSubscriptionBatchItem subscription,
        DateTime newPeriodStart,
        DateTime newPeriodEnd)
    {
        return JsonSerializer.Serialize(new
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
                periodStart = newPeriodStart,
                periodEnd = newPeriodEnd
            }
        });
    }
}
