using System.Diagnostics;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;
using UserService.Application.Interfaces;
using UserService.Domain.Enums;

namespace UserService.Infrastructure.Messaging;

public class SubscriptionSyncConsumer : BackgroundService
{
    private static readonly ActivitySource ActivitySource = new("UserService");
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly ILogger<SubscriptionSyncConsumer> _logger;
    private readonly string _connectionString;
    private readonly string _queueName;

    public SubscriptionSyncConsumer(
        IServiceScopeFactory scopeFactory,
        ILogger<SubscriptionSyncConsumer> logger,
        IConfiguration configuration)
    {
        _scopeFactory = scopeFactory;
        _logger = logger;
        _connectionString = configuration["RabbitMQ:ConnectionString"]
            ?? "amqp://guest:guest@localhost:5672/";
        _queueName = "user.subscription-sync";
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
                _logger.LogError(ex, "Subscription sync consumer error, reconnecting in 5 seconds...");
                await Task.Delay(TimeSpan.FromSeconds(5), stoppingToken);
            }
        }
    }

    private async Task ConnectAndConsumeAsync(CancellationToken stoppingToken)
    {
        var factory = new ConnectionFactory { Uri = new Uri(_connectionString) };

        await using var connection = await factory.CreateConnectionAsync(stoppingToken);
        await using var channel = await connection.CreateChannelAsync(cancellationToken: stoppingToken);

        await channel.BasicQosAsync(prefetchSize: 0, prefetchCount: 1, global: false, cancellationToken: stoppingToken);

        _logger.LogInformation("Connected to RabbitMQ, consuming from queue {QueueName}", _queueName);

        var consumer = new AsyncEventingBasicConsumer(channel);

        consumer.ReceivedAsync += async (_, ea) =>
        {
            // Extract trace context from AMQP headers
            ActivityContext parentContext = default;
            if (ea.BasicProperties.Headers?.TryGetValue("traceparent", out var tp) == true)
            {
                var traceparent = tp is byte[] tpBytes
                    ? Encoding.UTF8.GetString(tpBytes)
                    : tp?.ToString();
                if (traceparent is not null)
                    ActivityContext.TryParse(traceparent, null, out parentContext);
            }

            using var activity = ActivitySource.StartActivity(
                "subscription-sync.process",
                ActivityKind.Consumer,
                parentContext,
                tags: new[] { new KeyValuePair<string, object?>("messaging.routing_key", ea.RoutingKey) });

            try
            {
                var body = ea.Body.ToArray();
                var json = Encoding.UTF8.GetString(body);
                var routingKey = ea.RoutingKey;

                _logger.LogDebug("Received subscription event with routing key {RoutingKey}: {Json}", routingKey, json);

                await HandleMessageAsync(routingKey, json, stoppingToken);
                await channel.BasicAckAsync(ea.DeliveryTag, multiple: false, cancellationToken: stoppingToken);
            }
            catch (Exception ex)
            {
                activity?.SetStatus(ActivityStatusCode.Error, ex.Message);
                _logger.LogError(ex, "Error processing subscription sync message");
                await channel.BasicNackAsync(ea.DeliveryTag, multiple: false, requeue: true, cancellationToken: stoppingToken);
            }
        };

        await channel.BasicConsumeAsync(
            queue: _queueName,
            autoAck: false,
            consumer: consumer,
            cancellationToken: stoppingToken);

        var tcs = new TaskCompletionSource();
        stoppingToken.Register(() => tcs.TrySetResult());
        await tcs.Task;
    }

    private async Task HandleMessageAsync(string routingKey, string json, CancellationToken cancellationToken)
    {
        using var scope = _scopeFactory.CreateScope();
        var userRepository = scope.ServiceProvider.GetRequiredService<IUserRepository>();

        switch (routingKey)
        {
            case "subscription.created":
                var created = JsonSerializer.Deserialize<SubscriptionCreatedMessage>(json, JsonOptions);
                if (created is not null)
                {
                    await UpdateUserTierAsync(userRepository, created.UserId, created.Tier, cancellationToken);
                    _logger.LogInformation(
                        "Updated user {UserId} tier to {Tier} after subscription created",
                        created.UserId, created.Tier);
                }
                break;

            case "subscription.cancelled":
                var cancelled = JsonSerializer.Deserialize<SubscriptionCancelledMessage>(json, JsonOptions);
                if (cancelled is not null)
                {
                    await UpdateUserTierAsync(userRepository, cancelled.UserId, "Free", cancellationToken);
                    _logger.LogInformation(
                        "Updated user {UserId} tier to Free after subscription cancelled",
                        cancelled.UserId);
                }
                break;

            case "plan.changed":
                var changed = JsonSerializer.Deserialize<PlanChangedMessage>(json, JsonOptions);
                if (changed is not null)
                {
                    await UpdateUserTierAsync(userRepository, changed.UserId, changed.NewTier, cancellationToken);
                    _logger.LogInformation(
                        "Updated user {UserId} tier from {OldTier} to {NewTier} after plan changed",
                        changed.UserId, changed.OldTier, changed.NewTier);
                }
                break;

            default:
                _logger.LogWarning("Unknown routing key: {RoutingKey}", routingKey);
                break;
        }
    }

    private static async Task UpdateUserTierAsync(
        IUserRepository userRepository,
        Guid userId,
        string tierString,
        CancellationToken cancellationToken)
    {
        var user = await userRepository.GetByIdAsync(userId, cancellationToken);
        if (user is null)
        {
            return;
        }

        // Don't downgrade admin users
        if (user.Role == SubscriptionTier.Admin)
        {
            return;
        }

        if (Enum.TryParse<SubscriptionTier>(tierString, ignoreCase: true, out var tier))
        {
            user.Role = tier;
            user.UpdatedAt = DateTime.UtcNow;
            await userRepository.UpdateAsync(user, cancellationToken);
        }
    }

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };
}
