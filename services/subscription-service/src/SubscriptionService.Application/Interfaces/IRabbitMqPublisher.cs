namespace SubscriptionService.Application.Interfaces;

public interface IRabbitMqPublisher
{
    Task PublishAsync(string exchange, string routingKey, string payload, CancellationToken cancellationToken = default);
    Task<bool> IsConnectedAsync();
}
