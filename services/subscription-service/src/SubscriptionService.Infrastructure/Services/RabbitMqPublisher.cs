using System.Text;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using RabbitMQ.Client;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Infrastructure.Services;

public class RabbitMqPublisher : IRabbitMqPublisher, IAsyncDisposable
{
    private readonly ILogger<RabbitMqPublisher> _logger;
    private readonly string _connectionString;
    private IConnection? _connection;
    private IChannel? _channel;
    private readonly SemaphoreSlim _lock = new(1, 1);
    private readonly HashSet<string> _declaredExchanges = new();

    public RabbitMqPublisher(IConfiguration configuration, ILogger<RabbitMqPublisher> logger)
    {
        _logger = logger;
        _connectionString = configuration["RabbitMQ:ConnectionString"]
            ?? "amqp://guest:guest@localhost:5672/";
    }

    public async Task PublishAsync(string exchange, string routingKey, string payload, CancellationToken cancellationToken = default)
    {
        var channel = await GetChannelAsync(cancellationToken);

        if (!_declaredExchanges.Contains(exchange))
        {
            await channel.ExchangeDeclareAsync(
                exchange: exchange,
                type: ExchangeType.Topic,
                durable: true,
                autoDelete: false,
                cancellationToken: cancellationToken);
            _declaredExchanges.Add(exchange);
        }

        var properties = new BasicProperties
        {
            DeliveryMode = DeliveryModes.Persistent,
            ContentType = "application/json",
            Timestamp = new AmqpTimestamp(DateTimeOffset.UtcNow.ToUnixTimeSeconds())
        };

        var body = Encoding.UTF8.GetBytes(payload);

        await channel.BasicPublishAsync(
            exchange: exchange,
            routingKey: routingKey,
            mandatory: false,
            basicProperties: properties,
            body: body,
            cancellationToken: cancellationToken);

        _logger.LogDebug(
            "Published message to exchange '{Exchange}' with routing key '{RoutingKey}'",
            exchange, routingKey);
    }

    public async Task<bool> IsConnectedAsync()
    {
        try
        {
            if (_connection is null || !_connection.IsOpen)
                return false;

            var channel = await GetChannelAsync();
            return channel.IsOpen;
        }
        catch
        {
            return false;
        }
    }

    private async Task<IChannel> GetChannelAsync(CancellationToken cancellationToken = default)
    {
        if (_channel is { IsOpen: true })
            return _channel;

        await _lock.WaitAsync(cancellationToken);
        try
        {
            if (_channel is { IsOpen: true })
                return _channel;

            if (_connection is null || !_connection.IsOpen)
            {
                var factory = new ConnectionFactory { Uri = new Uri(_connectionString) };
                _connection = await factory.CreateConnectionAsync(cancellationToken);
                _declaredExchanges.Clear();
                _logger.LogInformation("RabbitMQ connection established.");
            }

            _channel = await _connection.CreateChannelAsync(cancellationToken: cancellationToken);
            _logger.LogInformation("RabbitMQ channel created.");
            return _channel;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to establish RabbitMQ connection or channel.");
            throw;
        }
        finally
        {
            _lock.Release();
        }
    }

    public async ValueTask DisposeAsync()
    {
        if (_channel is not null)
        {
            await _channel.CloseAsync();
            _channel.Dispose();
        }
        if (_connection is not null)
        {
            await _connection.CloseAsync();
            _connection.Dispose();
        }
        _lock.Dispose();
        GC.SuppressFinalize(this);
    }
}
