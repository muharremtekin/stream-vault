namespace SubscriptionService.Domain.Exceptions;

public class PaymentFailedException : Exception
{
    public PaymentFailedException()
        : base("Payment processing failed.") { }

    public PaymentFailedException(string reason)
        : base($"Payment processing failed: {reason}")
    {
        Reason = reason;
    }

    public PaymentFailedException(string message, Exception innerException)
        : base(message, innerException) { }

    public string? Reason { get; }
}
