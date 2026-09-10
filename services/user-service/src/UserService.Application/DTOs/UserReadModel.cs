using UserService.Domain.Enums;

namespace UserService.Application.DTOs;

public class UserReadModel
{
    public Guid Id { get; set; }

    public string Email { get; set; } = string.Empty;

    public string PasswordHash { get; set; } = string.Empty;

    public SubscriptionTier Role { get; set; }

    public int ProfileCount { get; set; }
}
