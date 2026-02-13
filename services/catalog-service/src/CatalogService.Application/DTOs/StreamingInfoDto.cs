namespace CatalogService.Application.DTOs;

public record StreamingInfoDto
{
    public string VideoStatus { get; init; } = string.Empty;
    public int? DurationSeconds { get; init; }
    public List<QualityInfoDto> AvailableQualities { get; init; } = new();
    public string? ManifestUrl { get; init; }
    public string? ThumbnailUrl { get; init; }
    public string? PosterUrl { get; init; }
    public DateTime? EncodedAt { get; init; }
}

public record QualityInfoDto
{
    public string Label { get; init; } = string.Empty;
    public int Width { get; init; }
    public int Height { get; init; }
    public int BitrateKbps { get; init; }
    public int SegmentCount { get; init; }
}
