using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using MongoDB.Bson.Serialization.Attributes;

namespace CatalogService.Domain.Entities;

public class Episode
{
    public int EpisodeNumber { get; set; }

    public string Title { get; set; } = string.Empty;

    public string Description { get; set; } = string.Empty;

    public Duration Duration { get; set; } = new();

    public string ThumbnailUrl { get; set; } = string.Empty;

    [BsonRepresentation(MongoDB.Bson.BsonType.String)]
    public VideoStatus VideoStatus { get; set; } = VideoStatus.NotUploaded;

    public StreamingInfo? StreamingInfo { get; set; }

    public Episode() { }

    public Episode(int episodeNumber, string title, string description, Duration duration, string thumbnailUrl)
    {
        EpisodeNumber = episodeNumber;
        Title = title ?? throw new ArgumentNullException(nameof(title));
        Description = description ?? throw new ArgumentNullException(nameof(description));
        Duration = duration ?? throw new ArgumentNullException(nameof(duration));
        ThumbnailUrl = thumbnailUrl ?? throw new ArgumentNullException(nameof(thumbnailUrl));
    }
}
