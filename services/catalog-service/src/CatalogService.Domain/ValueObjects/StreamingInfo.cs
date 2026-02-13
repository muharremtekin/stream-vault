namespace CatalogService.Domain.ValueObjects;

public class StreamingInfo
{
    public int DurationSeconds { get; private set; }
    public List<QualityInfo> AvailableQualities { get; private set; } = new();
    public string ManifestPath { get; private set; } = string.Empty;
    public string? ThumbnailPath { get; private set; }
    public string? PosterPath { get; private set; }
    public DateTime EncodedAt { get; private set; }

    public StreamingInfo() { }

    public StreamingInfo(
        int durationSeconds,
        List<QualityInfo> availableQualities,
        string manifestPath,
        string? thumbnailPath,
        string? posterPath,
        DateTime encodedAt)
    {
        DurationSeconds = durationSeconds;
        AvailableQualities = availableQualities ?? throw new ArgumentNullException(nameof(availableQualities));
        ManifestPath = manifestPath ?? throw new ArgumentNullException(nameof(manifestPath));
        ThumbnailPath = thumbnailPath;
        PosterPath = posterPath;
        EncodedAt = encodedAt;
    }
}
