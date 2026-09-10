using Microsoft.Extensions.Configuration;

namespace UserService.Infrastructure.Discovery;

public sealed record ConsulRegistrationSettings(string Name, string Host, int Port)
{
    public string HealthCheckAddress => $"http://{Host}:{Port}/health/ready";

    public static ConsulRegistrationSettings FromConfiguration(
        IConfiguration configuration,
        string defaultName,
        int defaultPort)
    {
        var name = configuration["ServiceRegistration:Name"] ?? defaultName;
        var host = configuration["Service:Host"] ?? "localhost";
        var port = configuration.GetValue<int?>("ServiceRegistration:Port") ?? defaultPort;

        if (string.IsNullOrWhiteSpace(name))
            throw new InvalidOperationException("ServiceRegistration:Name must not be empty.");
        if (string.IsNullOrWhiteSpace(host))
            throw new InvalidOperationException("Service:Host must not be empty.");
        if (port is < 1 or > 65535)
            throw new InvalidOperationException("ServiceRegistration:Port must be between 1 and 65535.");

        return new ConsulRegistrationSettings(name, host, port);
    }
}
