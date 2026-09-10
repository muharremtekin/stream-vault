using CatalogService.Infrastructure.Discovery;
using Microsoft.Extensions.Configuration;
using Xunit;

namespace CatalogService.UnitTests.Configuration;

public class ConsulRegistrationSettingsTests
{
    [Fact]
    public void CustomNameHostAndPortAreUsedByRegistrationAndHealthCheck()
    {
        var configuration = new ConfigurationBuilder().AddInMemoryCollection(
            new Dictionary<string, string?>
            {
                ["ServiceRegistration:Name"] = "custom-catalog",
                ["ServiceRegistration:Port"] = "15100",
                ["Service:Host"] = "catalog.internal"
            }).Build();

        var settings = ConsulRegistrationSettings.FromConfiguration(configuration, "catalog-service", 5100);
        Assert.Equal("custom-catalog", settings.Name);
        Assert.Equal("catalog.internal", settings.Host);
        Assert.Equal(15100, settings.Port);
        Assert.Equal("http://catalog.internal:15100/health/ready", settings.HealthCheckAddress);
    }
}
