namespace SubscriptionService.Domain.Exceptions;

public class SubscriptionNotFoundException : Exception
{
    public SubscriptionNotFoundException()
        : base("Subscription was not found.") { }

    public SubscriptionNotFoundException(Guid subscriptionId)
        : base($"Subscription with ID '{subscriptionId}' was not found.")
    {
        SubscriptionId = subscriptionId;
    }

    public SubscriptionNotFoundException(string message)
        : base(message) { }

    public SubscriptionNotFoundException(string message, Exception innerException)
        : base(message, innerException) { }

    public Guid? SubscriptionId { get; }
}
