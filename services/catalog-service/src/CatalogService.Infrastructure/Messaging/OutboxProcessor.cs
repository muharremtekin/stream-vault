using CatalogService.Application.Events;
using CatalogService.Application.Interfaces;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace CatalogService.Infrastructure.Messaging;

public class OutboxProcessor : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly IEventPublisher _publisher;
    private readonly ILogger<OutboxProcessor> _logger;
    private readonly TimeSpan _interval = TimeSpan.FromSeconds(5);
    private const int MaxRetryCount = 3;

    public OutboxProcessor(
        IServiceProvider serviceProvider,
        IEventPublisher publisher,
        ILogger<OutboxProcessor> logger)
    {
        _serviceProvider = serviceProvider;
        _publisher = publisher;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Catalog outbox processor started.");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessOutboxMessagesAsync(stoppingToken);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error processing outbox messages.");
            }

            await Task.Delay(_interval, stoppingToken);
        }

        _logger.LogInformation("Catalog outbox processor stopped.");
    }

    private async Task ProcessOutboxMessagesAsync(CancellationToken cancellationToken)
    {
        using var scope = _serviceProvider.CreateScope();
        var outboxRepository = scope.ServiceProvider.GetRequiredService<IOutboxRepository>();

        var messages = await outboxRepository.GetUnprocessedAsync(50, cancellationToken);
        if (messages.Count == 0)
            return;

        _logger.LogDebug("Processing {Count} outbox messages.", messages.Count);

        foreach (var message in messages)
        {
            if (message.RetryCount >= MaxRetryCount)
            {
                _logger.LogError(
                    "Outbox message {MessageId} exceeded max retry count ({MaxRetry}). EventType: {EventType}",
                    message.Id, MaxRetryCount, message.EventType);
                continue;
            }

            try
            {
                await _publisher.PublishAsync(
                    message.EventType,
                    message.Payload,
                    cancellationToken);

                await outboxRepository.MarkAsProcessedAsync(message.Id, cancellationToken);

                _logger.LogInformation(
                    "Published outbox message {MessageId} with routing key '{RoutingKey}'.",
                    message.Id, message.EventType);
            }
            catch (Exception ex)
            {
                _logger.LogWarning(ex,
                    "Failed to publish outbox message {MessageId}. Retry {Retry}/{MaxRetry}.",
                    message.Id, message.RetryCount + 1, MaxRetryCount);

                await outboxRepository.IncrementRetryAsync(message.Id, ex.Message, cancellationToken);
            }
        }
    }
}
