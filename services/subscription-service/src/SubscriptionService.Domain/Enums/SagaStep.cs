namespace SubscriptionService.Domain.Enums;

public enum SagaStep
{
    // Subscription Creation Saga
    ValidatePlan = 0,
    CreateSubscription = 1,
    ProcessPayment = 2,
    Activate = 3,
    CreateInvoice = 4,
    PublishEvents = 5,

    // Plan Change Saga
    ValidateChange = 10,
    ProcessPriceDifference = 11,
    UpdateSubscription = 12,
    PublishChangeEvents = 13
}
