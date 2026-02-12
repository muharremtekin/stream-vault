using UserService.Domain.Enums;

namespace UserService.Domain.Entities;

public class Profile
{
    public Guid Id { get; set; } = Guid.NewGuid();

    public Guid UserId { get; set; }

    public string Name { get; set; } = string.Empty;

    public ProfileIcon Icon { get; set; } = ProfileIcon.Avatar1;

    public bool IsKids { get; set; }

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public User User { get; set; } = null!;

    public List<WatchlistItem> WatchlistItems { get; set; } = new();
}
