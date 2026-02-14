namespace UserService.Domain.Entities;

public class ContentRating
{
    public Guid Id { get; set; } = Guid.NewGuid();

    public Guid UserId { get; set; }

    public string ContentId { get; set; } = string.Empty;

    public int Rating { get; set; }

    public DateTime RatedAt { get; set; } = DateTime.UtcNow;

    public User User { get; set; } = null!;
}
