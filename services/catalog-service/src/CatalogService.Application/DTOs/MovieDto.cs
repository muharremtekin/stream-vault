using CatalogService.Domain.Enums;

namespace CatalogService.Application.DTOs;

public record MovieDto
{
    public string Id { get; init; } = string.Empty;
    public string Title { get; init; } = string.Empty;
    public string? OriginalTitle { get; init; }
    public string Description { get; init; } = string.Empty;
    public int ReleaseYear { get; init; }
    public int DurationMinutes { get; init; }
    public string DurationFormatted { get; init; } = string.Empty;
    public string MaturityRating { get; init; } = string.Empty;
    public List<string> Genres { get; init; } = new();
    public List<CastMemberDto> Cast { get; init; } = new();
    public string Director { get; init; } = string.Empty;
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string BannerUrl { get; init; } = string.Empty;
    public string? TrailerUrl { get; init; }
    public double AverageRating { get; init; }
    public int RatingCount { get; init; }
    public string Status { get; init; } = string.Empty;
    public List<string> Tags { get; init; } = new();
    public DateTime CreatedAt { get; init; }
    public DateTime UpdatedAt { get; init; }
}

public record CastMemberDto
{
    public string Name { get; init; } = string.Empty;
    public string Role { get; init; } = string.Empty;
    public string? PhotoUrl { get; init; }
}
