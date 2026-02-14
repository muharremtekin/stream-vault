using System.Text;
using System.Text.Json;
using CatalogService.Application.Commands.UpdateVideoStatus;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using MediatR;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;

namespace CatalogService.Infrastructure.Messaging;

public class EncodingResultConsumer : BackgroundService
{
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly ILogger<EncodingResultConsumer> _logger;
    private readonly string _connectionString;
    private readonly string _queueName;

    public EncodingResultConsumer(
        IServiceScopeFactory scopeFactory,
        ILogger<EncodingResultConsumer> logger,
        IConfiguration configuration)
    {
        _scopeFactory = scopeFactory;
        _logger = logger;
        _connectionString = configuration.GetValue<string>("RabbitMQ:ConnectionString")
            ?? throw new InvalidOperationException("RabbitMQ:ConnectionString configuration is required");
        _queueName = configuration.GetValue<string>("RabbitMQ:QueueName")
            ?? "encoding.results.catalog";
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ConnectAndConsumeAsync(stoppingToken);
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "RabbitMQ consumer error, reconnecting in 5 seconds...");
                await Task.Delay(TimeSpan.FromSeconds(5), stoppingToken);
            }
        }
    }

    private async Task ConnectAndConsumeAsync(CancellationToken stoppingToken)
    {
        var factory = new ConnectionFactory
        {
            Uri = new Uri(_connectionString)
        };

        await using var connection = await factory.CreateConnectionAsync(stoppingToken);
        await using var channel = await connection.CreateChannelAsync(cancellationToken: stoppingToken);

        await channel.BasicQosAsync(prefetchSize: 0, prefetchCount: 1, global: false, cancellationToken: stoppingToken);

        _logger.LogInformation("Connected to RabbitMQ, consuming from queue {QueueName}", _queueName);

        var consumer = new AsyncEventingBasicConsumer(channel);

        consumer.ReceivedAsync += async (_, ea) =>
        {
            try
            {
                var body = ea.Body.ToArray();
                var json = Encoding.UTF8.GetString(body);

                _logger.LogDebug("Received encoding result: {Json}", json);

                var message = JsonSerializer.Deserialize<EncodingResultMessage>(json);
                if (message is null)
                {
                    _logger.LogWarning("Failed to deserialize encoding result message");
                    await channel.BasicNackAsync(ea.DeliveryTag, multiple: false, requeue: false, cancellationToken: stoppingToken);
                    return;
                }

                await HandleMessageAsync(message, stoppingToken);
                await channel.BasicAckAsync(ea.DeliveryTag, multiple: false, cancellationToken: stoppingToken);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error processing encoding result message");
                await channel.BasicNackAsync(ea.DeliveryTag, multiple: false, requeue: true, cancellationToken: stoppingToken);
            }
        };

        await channel.BasicConsumeAsync(
            queue: _queueName,
            autoAck: false,
            consumer: consumer,
            cancellationToken: stoppingToken);

        // Keep alive until cancelled
        var tcs = new TaskCompletionSource();
        stoppingToken.Register(() => tcs.TrySetResult());
        await tcs.Task;
    }

    internal async Task HandleMessageAsync(EncodingResultMessage message, CancellationToken cancellationToken)
    {
        _logger.LogInformation(
            "Processing encoding result: EventId={EventId}, EventType={EventType}, Source={Source}, " +
            "CorrelationId={CorrelationId}, JobId={JobId}, ContentId={ContentId}, Status={Status}",
            message.EventId, message.EventType, message.Source,
            message.CorrelationId, message.JobId, message.ContentId, message.Status);

        using var scope = _scopeFactory.CreateScope();
        var mediator = scope.ServiceProvider.GetRequiredService<IMediator>();

        VideoStatus videoStatus;
        StreamingInfo? streamingInfo = null;

        if (message.Status == "completed")
        {
            videoStatus = VideoStatus.Ready;

            var manifestPath = $"{message.ContentId}/manifest.m3u8";

            var qualities = message.Outputs.Select(o => new QualityInfo(
                label: o.Quality,
                width: o.Width,
                height: o.Height,
                bitrateKbps: o.BitrateKbps,
                segmentCount: o.SegmentCount
            )).ToList();

            var encodedAt = DateTime.TryParse(message.CompletedAt, out var parsed)
                ? parsed.ToUniversalTime()
                : DateTime.UtcNow;

            streamingInfo = new StreamingInfo(
                durationSeconds: (int)message.DurationSeconds,
                availableQualities: qualities,
                manifestPath: manifestPath,
                thumbnailPath: $"{message.ContentId}/thumb_300x170.jpg",
                posterPath: $"{message.ContentId}/poster.jpg",
                encodedAt: encodedAt
            );
        }
        else if (message.Status == "failed")
        {
            videoStatus = VideoStatus.Error;
            _logger.LogWarning(
                "Encoding failed for ContentId={ContentId}, CorrelationId={CorrelationId}: {Error}",
                message.ContentId, message.CorrelationId, message.ErrorMessage);
        }
        else
        {
            _logger.LogWarning("Unknown encoding result status: {Status}", message.Status);
            return;
        }

        var command = new UpdateVideoStatusCommand
        {
            ContentId = message.ContentId,
            VideoStatus = videoStatus,
            StreamingInfo = streamingInfo
        };

        var result = await mediator.Send(command, cancellationToken);

        if (!result)
        {
            _logger.LogWarning(
                "Failed to update video status for ContentId={ContentId} (movie not found or invalid transition)",
                message.ContentId);
        }
    }
}
