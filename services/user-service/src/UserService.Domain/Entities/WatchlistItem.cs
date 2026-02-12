using UserService.Domain.Enums;

namespace UserService.Domain.Entities;

public class WatchlistItem
{
    public Guid Id { get; set; } = Guid.NewGuid();

    public Guid ProfileId { get; set; }

    public string ContentId { get; set; } = string.Empty;

    public ContentType ContentType { get; set; }

    public DateTime AddedAt { get; set; } = DateTime.UtcNow;

    public string? Note { get; set; }

    public Profile Profile { get; set; } = null!;
}
