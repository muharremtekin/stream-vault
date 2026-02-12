using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Commands.CreateMovie;

public record CreateMovieCommand : IRequest<string>
{
    public string Title { get; init; } = string.Empty;
    public string? OriginalTitle { get; init; }
    public string Description { get; init; } = string.Empty;
    public int ReleaseYear { get; init; }
    public int DurationMinutes { get; init; }
    public MaturityRating MaturityRating { get; init; }
    public List<string> Genres { get; init; } = new();
    public List<CastMemberInput> Cast { get; init; } = new();
    public string Director { get; init; } = string.Empty;
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string BannerUrl { get; init; } = string.Empty;
    public string? TrailerUrl { get; init; }
    public List<string> Tags { get; init; } = new();
}

public record CastMemberInput(string Name, string Role, string? PhotoUrl = null);
