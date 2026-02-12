using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Commands.CreateSeries;

public class CreateSeriesHandler : IRequestHandler<CreateSeriesCommand, string>
{
    private readonly ISeriesRepository _seriesRepository;

    public CreateSeriesHandler(ISeriesRepository seriesRepository)
    {
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
    }

    public async Task<string> Handle(CreateSeriesCommand request, CancellationToken cancellationToken)
    {
        var series = new Series
        {
            Title = request.Title,
            Description = request.Description,
            ReleaseYear = request.ReleaseYear,
            MaturityRating = request.MaturityRating,
            Genres = request.Genres,
            Cast = request.Cast.Select(c => new CastMember(c.Name, c.Role, c.PhotoUrl)).ToList(),
            Creator = request.Creator,
            ThumbnailUrl = request.ThumbnailUrl,
            BannerUrl = request.BannerUrl,
            Status = ContentStatus.Draft,
            Seasons = new List<Season>(),
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };

        await _seriesRepository.AddAsync(series, cancellationToken);

        return series.Id;
    }
}
