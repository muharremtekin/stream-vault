namespace UserService.Application.DTOs;

public class WatchlistItemDto
{
    public Guid Id { get; set; }

    public string ContentId { get; set; } = string.Empty;

    public string ContentType { get; set; } = string.Empty;

    public DateTime AddedAt { get; set; }

    public string? Note { get; set; }
}
