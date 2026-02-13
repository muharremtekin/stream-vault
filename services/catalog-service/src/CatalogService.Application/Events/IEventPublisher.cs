namespace CatalogService.Application.Events;

public interface IEventPublisher
{
    Task PublishAsync<T>(string routingKey, T message, CancellationToken cancellationToken = default)
        where T : class;
}
