using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.ValueObjects;
using MediatR;

namespace CatalogService.Application.Commands.AddEpisode;

public class AddEpisodeHandler : IRequestHandler<AddEpisodeCommand, bool>
{
    private readonly ISeriesRepository _seriesRepository;

    public AddEpisodeHandler(ISeriesRepository seriesRepository)
    {
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
    }

    public async Task<bool> Handle(AddEpisodeCommand request, CancellationToken cancellationToken)
    {
        var series = await _seriesRepository.GetByIdAsync(request.SeriesId, cancellationToken);

        if (series is null)
            throw new KeyNotFoundException($"Series with ID '{request.SeriesId}' was not found.");

        var season = series.GetSeason(request.SeasonNumber);

        if (season is null)
        {
            season = new Season(request.SeasonNumber, series.ReleaseYear);
            series.AddSeason(season);
        }

        var episode = new Episode(
            request.EpisodeNumber,
            request.Title,
            request.Description,
            Duration.FromMinutes(request.DurationMinutes),
            request.ThumbnailUrl
        );

        season.AddEpisode(episode);
        series.UpdatedAt = DateTime.UtcNow;

        await _seriesRepository.UpdateAsync(series, cancellationToken);

        return true;
    }
}
