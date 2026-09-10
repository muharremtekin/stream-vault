using Microsoft.Extensions.Configuration;
using UserService.Infrastructure.Discovery;
using Xunit;

namespace UserService.UnitTests.Configuration;

public class ConsulRegistrationSettingsTests
{
    [Fact]
    public void CustomNameHostAndPortAreUsedByRegistrationAndHealthCheck()
    {
        var configuration = new ConfigurationBuilder().AddInMemoryCollection(
            new Dictionary<string, string?>
            {
                ["ServiceRegistration:Name"] = "custom-user",
                ["ServiceRegistration:Port"] = "18180",
                ["Service:Host"] = "user.internal"
            }).Build();

        var settings = ConsulRegistrationSettings.FromConfiguration(configuration, "user-service", 8080);
        Assert.Equal("custom-user", settings.Name);
        Assert.Equal("user.internal", settings.Host);
        Assert.Equal(18180, settings.Port);
        Assert.Equal("http://user.internal:18180/health/ready", settings.HealthCheckAddress);
    }
}
