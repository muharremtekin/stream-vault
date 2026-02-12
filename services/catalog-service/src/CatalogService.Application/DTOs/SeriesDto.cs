namespace CatalogService.Application.DTOs;

public record SeriesDto
{
    public string Id { get; init; } = string.Empty;
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int ReleaseYear { get; init; }
    public string MaturityRating { get; init; } = string.Empty;
    public List<string> Genres { get; init; } = new();
    public List<CastMemberDto> Cast { get; init; } = new();
    public string Creator { get; init; } = string.Empty;
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string BannerUrl { get; init; } = string.Empty;
    public string Status { get; init; } = string.Empty;
    public List<SeasonDto> Seasons { get; init; } = new();
    public int TotalSeasons { get; init; }
    public int TotalEpisodes { get; init; }
    public DateTime CreatedAt { get; init; }
    public DateTime UpdatedAt { get; init; }
}

public record SeasonDto
{
    public int SeasonNumber { get; init; }
    public string? Title { get; init; }
    public int ReleaseYear { get; init; }
    public List<EpisodeDto> Episodes { get; init; } = new();
    public int EpisodeCount { get; init; }
}
