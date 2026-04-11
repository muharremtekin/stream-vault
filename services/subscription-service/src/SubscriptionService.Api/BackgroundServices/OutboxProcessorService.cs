using System.Diagnostics.Metrics;
using Microsoft.Extensions.Options;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Api.BackgroundServices;

public class OutboxProcessorService : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly IRabbitMqPublisher _publisher;
    private readonly ILogger<OutboxProcessorService> _logger;
    private readonly OutboxProcessorOptions _options;

    private static readonly Meter Meter = new("SubscriptionService.Outbox");
    private static readonly Counter<long> PublishedCounter = Meter.CreateCounter<long>("outbox.messages.published");
    private static readonly Counter<long> FailedCounter = Meter.CreateCounter<long>("outbox.messages.failed");
    private static readonly Counter<long> DeadLetterCounter = Meter.CreateCounter<long>("outbox.messages.dead_lettered");

    private const string Exchange = "subscription.events";

    public OutboxProcessorService(
        IServiceProvider serviceProvider,
        IRabbitMqPublisher publisher,
        ILogger<OutboxProcessorService> logger,
        IOptions<OutboxProcessorOptions> options)
    {
        _serviceProvider = serviceProvider;
        _publisher = publisher;
        _logger = logger;
        _options = options.Value;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation(
            "Outbox processor started. Interval={Interval}s, MaxRetry={MaxRetry}, BatchSize={BatchSize}",
            _options.IntervalSeconds, _options.MaxRetryCount, _options.BatchSize);

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                if (!await _publisher.IsConnectedAsync())
                {
                    _logger.LogWarning("RabbitMQ connection is not available. Skipping outbox processing cycle.");
                }
                else
                {
                    await ProcessOutboxMessagesAsync(stoppingToken);
                }
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Unexpected error during outbox processing cycle.");
            }

            try
            {
                await Task.Delay(TimeSpan.FromSeconds(_options.IntervalSeconds), stoppingToken);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
        }

        _logger.LogInformation("Shutdown requested. Processing remaining outbox messages...");
        using var shutdownCts = new CancellationTokenSource(TimeSpan.FromSeconds(_options.ShutdownTimeoutSeconds));
        try
        {
            if (await _publisher.IsConnectedAsync())
            {
                await ProcessOutboxMessagesAsync(shutdownCts.Token);
                _logger.LogInformation("Remaining outbox messages processed successfully.");
            }
            else
            {
                _logger.LogWarning("RabbitMQ connection is not available during shutdown. Skipping drain.");
            }
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Failed to process remaining outbox messages during shutdown.");
        }

        _logger.LogInformation("Outbox processor stopped.");
    }

    private async Task ProcessOutboxMessagesAsync(CancellationToken cancellationToken)
    {
        using var scope = _serviceProvider.CreateScope();
        var outboxRepository = scope.ServiceProvider.GetRequiredService<IOutboxRepository>();
        var totalMessagesProcessed = 0;

        while (true)
        {
            cancellationToken.ThrowIfCancellationRequested();

            var messages = await outboxRepository.GetUnprocessedAsync(_options.BatchSize, cancellationToken);
            if (messages.Count == 0)
            {
                break;
            }

            _logger.LogDebug("Processing {Count} outbox messages.", messages.Count);

            foreach (var message in messages)
            {
                cancellationToken.ThrowIfCancellationRequested();

                if (message.RetryCount >= _options.MaxRetryCount)
                {
                    _logger.LogError(
                        "Outbox message {MessageId} exceeded max retry count ({MaxRetry}). EventType: {EventType}. Moving to dead-letter.",
                        message.Id, _options.MaxRetryCount, message.EventType);

                    await outboxRepository.MarkAsDeadLetterAsync(message.Id, cancellationToken);
                    DeadLetterCounter.Add(1);
                    totalMessagesProcessed++;
                    continue;
                }

                try
                {
                    await _publisher.PublishAsync(
                        Exchange,
                        message.EventType,
                        message.Payload,
                        cancellationToken);

                    await outboxRepository.MarkAsProcessedAsync(message.Id, cancellationToken);

                    PublishedCounter.Add(1);
                    totalMessagesProcessed++;
                    _logger.LogInformation(
                        "Published outbox message {MessageId} to '{Exchange}' with routing key '{RoutingKey}'.",
                        message.Id, Exchange, message.EventType);
                }
                catch (Exception ex)
                {
                    FailedCounter.Add(1);
                    totalMessagesProcessed++;
                    _logger.LogWarning(ex,
                        "Failed to publish outbox message {MessageId}. Retry {Retry}/{MaxRetry}.",
                        message.Id, message.RetryCount + 1, _options.MaxRetryCount);

                    await outboxRepository.IncrementRetryAsync(message.Id, ex.Message, cancellationToken);
                }
            }

            if (messages.Count < _options.BatchSize)
            {
                break;
            }
        }

        if (totalMessagesProcessed > 0)
        {
            _logger.LogDebug(
                "Completed outbox processing cycle after handling {Count} due messages.",
                totalMessagesProcessed);
        }
    }
}
