using Microsoft.Extensions.Diagnostics.HealthChecks;
using RabbitMQ.Client;

namespace SubscriptionService.Api.Health;

internal sealed class RabbitMqHealthCheck : IHealthCheck, IAsyncDisposable
{
    private readonly string _connectionString;
    private readonly SemaphoreSlim _connectionLock = new(1, 1);
    private IConnection? _connection;

    public RabbitMqHealthCheck(IConfiguration configuration)
    {
        _connectionString = configuration["RabbitMQ:ConnectionString"]
            ?? "amqp://guest:guest@localhost:5672/";
    }

    public async Task<HealthCheckResult> CheckHealthAsync(
        HealthCheckContext context,
        CancellationToken cancellationToken = default)
    {
        try
        {
            if (_connection is { IsOpen: true })
                return HealthCheckResult.Healthy();

            await _connectionLock.WaitAsync(cancellationToken);
            try
            {
                if (_connection is not { IsOpen: true })
                {
                    if (_connection is not null)
                        await _connection.DisposeAsync();

                    var factory = new ConnectionFactory { Uri = new Uri(_connectionString) };
                    _connection = await factory.CreateConnectionAsync(cancellationToken);
                }
            }
            finally
            {
                _connectionLock.Release();
            }

            return _connection.IsOpen
                ? HealthCheckResult.Healthy()
                : HealthCheckResult.Unhealthy("RabbitMQ connection is closed.");
        }
        catch (Exception exception)
        {
            return HealthCheckResult.Unhealthy("RabbitMQ connection failed.", exception);
        }
    }

    public async ValueTask DisposeAsync()
    {
        if (_connection is not null)
            await _connection.DisposeAsync();

        _connectionLock.Dispose();
    }
}
