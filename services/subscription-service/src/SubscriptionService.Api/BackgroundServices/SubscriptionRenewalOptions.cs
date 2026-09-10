namespace SubscriptionService.Api.BackgroundServices;

public class SubscriptionRenewalOptions
{
    public const string SectionName = "SubscriptionRenewal";

    public int IntervalSeconds { get; set; } = 3600;

    public int BatchSize { get; set; } = 100;
}
