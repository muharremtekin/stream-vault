namespace CatalogService.Application.DTOs;

public record EpisodeDto
{
    public int EpisodeNumber { get; init; }
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int DurationMinutes { get; init; }
    public string DurationFormatted { get; init; } = string.Empty;
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string VideoStatus { get; init; } = string.Empty;
}
