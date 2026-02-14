using System.Text;
using System.Text.Json;
using CatalogService.Application.Events;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using RabbitMQ.Client;

namespace CatalogService.Infrastructure.Messaging;

public class RabbitMqPublisher : IEventPublisher, IAsyncDisposable
{
    private readonly ILogger<RabbitMqPublisher> _logger;
    private readonly string _connectionString;
    private readonly string _exchange;
    private IConnection? _connection;
    private IChannel? _channel;
    private readonly SemaphoreSlim _lock = new(1, 1);
    private bool _exchangeDeclared;

    public RabbitMqPublisher(IConfiguration configuration, ILogger<RabbitMqPublisher> logger)
    {
        _logger = logger;
        _connectionString = configuration["RabbitMQ:ConnectionString"]
            ?? "amqp://guest:guest@localhost:5672/";
        _exchange = "catalog.events";
    }

    public async Task PublishAsync<T>(string routingKey, T message, CancellationToken cancellationToken = default)
        where T : class
    {
        var channel = await GetChannelAsync(cancellationToken);

        if (!_exchangeDeclared)
        {
            await channel.ExchangeDeclareAsync(
                exchange: _exchange,
                type: ExchangeType.Topic,
                durable: true,
                autoDelete: false,
                cancellationToken: cancellationToken);
            _exchangeDeclared = true;
        }

        var payload = JsonSerializer.Serialize(message);
        var body = Encoding.UTF8.GetBytes(payload);

        var properties = new BasicProperties
        {
            DeliveryMode = DeliveryModes.Persistent,
            ContentType = "application/json",
            Timestamp = new AmqpTimestamp(DateTimeOffset.UtcNow.ToUnixTimeSeconds())
        };

        await channel.BasicPublishAsync(
            exchange: _exchange,
            routingKey: routingKey,
            mandatory: false,
            basicProperties: properties,
            body: body,
            cancellationToken: cancellationToken);

        _logger.LogDebug(
            "Published message to exchange '{Exchange}' with routing key '{RoutingKey}'",
            _exchange, routingKey);
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
                _exchangeDeclared = false;
                _logger.LogInformation("RabbitMQ connection established for catalog events.");
            }

            _channel = await _connection.CreateChannelAsync(cancellationToken: cancellationToken);
            _logger.LogInformation("RabbitMQ channel created for catalog events.");
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
