using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace CatalogService.Domain.Entities;

public class Movie
{
    [BsonId]
    [BsonRepresentation(BsonType.ObjectId)]
    public string Id { get; set; } = ObjectId.GenerateNewId().ToString();

    public string Title { get; set; } = string.Empty;

    public string? OriginalTitle { get; set; }

    public string Description { get; set; } = string.Empty;

    public int ReleaseYear { get; set; }

    public Duration Duration { get; set; } = new();

    [BsonRepresentation(BsonType.String)]
    public MaturityRating MaturityRating { get; set; }

    public List<string> Genres { get; set; } = new();

    public List<CastMember> Cast { get; set; } = new();

    public string Director { get; set; } = string.Empty;

    public string ThumbnailUrl { get; set; } = string.Empty;

    public string BannerUrl { get; set; } = string.Empty;

    public string? TrailerUrl { get; set; }

    public double AverageRating { get; set; }

    public int RatingCount { get; set; }

    [BsonRepresentation(BsonType.String)]
    public ContentStatus Status { get; set; } = ContentStatus.Draft;

    public List<string> Tags { get; set; } = new();

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
