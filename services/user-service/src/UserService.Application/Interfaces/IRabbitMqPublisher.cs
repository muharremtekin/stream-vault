namespace UserService.Application.Interfaces;

public interface IRabbitMqPublisher
{
    Task PublishAsync(string exchange, string routingKey, string payload, CancellationToken cancellationToken = default);
}
