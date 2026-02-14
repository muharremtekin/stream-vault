namespace SubscriptionService.Domain.Exceptions;

public class PlanNotFoundException : Exception
{
    public PlanNotFoundException()
        : base("Plan was not found.") { }

    public PlanNotFoundException(Guid planId)
        : base($"Plan with ID '{planId}' was not found.")
    {
        PlanId = planId;
    }

    public PlanNotFoundException(string message)
        : base(message) { }

    public PlanNotFoundException(string message, Exception innerException)
        : base(message, innerException) { }

    public Guid? PlanId { get; }
}
