using CatalogService.Domain.Enums;

namespace CatalogService.Application.DTOs;

public record ContentSummaryDto
{
    public string Id { get; init; } = string.Empty;
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int ReleaseYear { get; init; }
    public ContentType ContentType { get; init; }
    public string MaturityRating { get; init; } = string.Empty;
    public List<string> Genres { get; init; } = new();
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string BannerUrl { get; init; } = string.Empty;
    public double? AverageRating { get; init; }
    public string Status { get; init; } = string.Empty;
}
