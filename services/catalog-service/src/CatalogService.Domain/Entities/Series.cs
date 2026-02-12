using CatalogService.Domain.Enums;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace CatalogService.Domain.Entities;

public class Series
{
    [BsonId]
    [BsonRepresentation(BsonType.ObjectId)]
    public string Id { get; set; } = ObjectId.GenerateNewId().ToString();

    public string Title { get; set; } = string.Empty;

    public string Description { get; set; } = string.Empty;

    public int ReleaseYear { get; set; }

    [BsonRepresentation(BsonType.String)]
    public MaturityRating MaturityRating { get; set; }

    public List<string> Genres { get; set; } = new();

    public List<CastMember> Cast { get; set; } = new();

    public string Creator { get; set; } = string.Empty;

    public string ThumbnailUrl { get; set; } = string.Empty;

    public string BannerUrl { get; set; } = string.Empty;

    [BsonRepresentation(BsonType.String)]
    public ContentStatus Status { get; set; } = ContentStatus.Draft;

    public List<Season> Seasons { get; set; } = new();

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public void AddSeason(Season season)
    {
        if (Seasons.Any(s => s.SeasonNumber == season.SeasonNumber))
            throw new InvalidOperationException(
                $"Season {season.SeasonNumber} already exists for series '{Title}'.");

        Seasons.Add(season);
        UpdatedAt = DateTime.UtcNow;
    }

    public Season? GetSeason(int seasonNumber)
    {
        return Seasons.FirstOrDefault(s => s.SeasonNumber == seasonNumber);
    }
}
