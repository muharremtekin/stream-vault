using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Commands.CreateSeries;

public record CreateSeriesCommand : IRequest<string>
{
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int ReleaseYear { get; init; }
    public MaturityRating MaturityRating { get; init; }
    public List<string> Genres { get; init; } = new();
    public List<CastMemberInput> Cast { get; init; } = new();
    public string Creator { get; init; } = string.Empty;
    public string ThumbnailUrl { get; init; } = string.Empty;
    public string BannerUrl { get; init; } = string.Empty;
}
