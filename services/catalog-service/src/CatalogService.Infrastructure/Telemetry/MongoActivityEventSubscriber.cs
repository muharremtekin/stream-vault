using System.Collections.Concurrent;
using System.Diagnostics;
using MongoDB.Driver.Core.Events;

namespace CatalogService.Infrastructure.Telemetry;

/// <summary>
/// Subscribes to MongoDB command events and creates Activity spans for OpenTelemetry.
/// </summary>
public sealed class MongoActivityEventSubscriber : IEventSubscriber
{
    public const string ActivitySourceName = "MongoDB.Driver";

    private static readonly ActivitySource Source = new(ActivitySourceName, "1.0.0");
    private readonly ConcurrentDictionary<int, Activity> _activities = new();

    private readonly ReflectionEventSubscriber _subscriber;

    public MongoActivityEventSubscriber()
    {
        _subscriber = new ReflectionEventSubscriber(this,
            bindingFlags: System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.NonPublic);
    }

    public bool TryGetEventHandler<TEvent>(out Action<TEvent> handler)
        => _subscriber.TryGetEventHandler(out handler);

    private void Handle(CommandStartedEvent @event)
    {
        var activity = Source.StartActivity($"mongodb.{@event.CommandName}", ActivityKind.Client);
        if (activity is null) return;

        activity.SetTag("db.system", "mongodb");
        activity.SetTag("db.name", @event.DatabaseNamespace?.DatabaseName);
        activity.SetTag("db.operation", @event.CommandName);

        _activities.TryAdd(@event.RequestId, activity);
    }

    private void Handle(CommandSucceededEvent @event)
    {
        if (_activities.TryRemove(@event.RequestId, out var activity))
        {
            activity.SetStatus(ActivityStatusCode.Ok);
            activity.Dispose();
        }
    }

    private void Handle(CommandFailedEvent @event)
    {
        if (_activities.TryRemove(@event.RequestId, out var activity))
        {
            activity.SetStatus(ActivityStatusCode.Error, @event.Failure?.Message);
            activity.Dispose();
        }
    }
}
