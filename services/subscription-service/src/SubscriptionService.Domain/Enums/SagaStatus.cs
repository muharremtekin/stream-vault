namespace SubscriptionService.Domain.Enums;

public enum SagaStatus
{
    Started = 0,
    InProgress = 1,
    Completed = 2,
    Compensating = 3,
    Failed = 4
}
