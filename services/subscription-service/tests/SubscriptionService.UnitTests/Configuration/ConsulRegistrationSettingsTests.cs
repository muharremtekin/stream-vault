using Microsoft.Extensions.Configuration;
using SubscriptionService.Infrastructure.Discovery;
using Xunit;

namespace SubscriptionService.UnitTests.Configuration;

public class ConsulRegistrationSettingsTests
{
    [Fact]
    public void CustomNameHostAndPortAreUsedByRegistrationAndHealthCheck()
    {
        var configuration = new ConfigurationBuilder().AddInMemoryCollection(
            new Dictionary<string, string?>
            {
                ["ServiceRegistration:Name"] = "custom-subscription",
                ["ServiceRegistration:Port"] = "15007",
                ["Service:Host"] = "subscription.internal"
            }).Build();

        var settings = ConsulRegistrationSettings.FromConfiguration(configuration, "subscription-service", 5007);
        Assert.Equal("custom-subscription", settings.Name);
        Assert.Equal("subscription.internal", settings.Host);
        Assert.Equal(15007, settings.Port);
        Assert.Equal("http://subscription.internal:15007/health/ready", settings.HealthCheckAddress);
    }
}
