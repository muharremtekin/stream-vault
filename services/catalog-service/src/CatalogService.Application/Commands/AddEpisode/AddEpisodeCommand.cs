using MediatR;

namespace CatalogService.Application.Commands.AddEpisode;

public record AddEpisodeCommand : IRequest<bool>
{
    public string SeriesId { get; init; } = string.Empty;
    public int SeasonNumber { get; init; }
    public int EpisodeNumber { get; init; }
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int DurationMinutes { get; init; }
    public string ThumbnailUrl { get; init; } = string.Empty;
}
