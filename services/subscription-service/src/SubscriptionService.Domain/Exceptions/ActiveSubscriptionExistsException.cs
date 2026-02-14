namespace SubscriptionService.Domain.Exceptions;

public class ActiveSubscriptionExistsException : Exception
{
    public ActiveSubscriptionExistsException()
        : base("User already has an active subscription.") { }

    public ActiveSubscriptionExistsException(Guid userId)
        : base($"User '{userId}' already has an active subscription.")
    {
        UserId = userId;
    }

    public ActiveSubscriptionExistsException(string message)
        : base(message) { }

    public ActiveSubscriptionExistsException(string message, Exception innerException)
        : base(message, innerException) { }

    public Guid? UserId { get; }
}
