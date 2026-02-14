using System.Text.Json;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.ValueObjects;
using MediatR;

namespace CatalogService.Application.Commands.AddEpisode;

public class AddEpisodeHandler : IRequestHandler<AddEpisodeCommand, bool>
{
    private readonly ISeriesRepository _seriesRepository;
    private readonly IOutboxRepository _outboxRepository;

    public AddEpisodeHandler(ISeriesRepository seriesRepository, IOutboxRepository outboxRepository)
    {
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _outboxRepository = outboxRepository ?? throw new ArgumentNullException(nameof(outboxRepository));
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

        var outboxMessage = new OutboxMessage
        {
            EventType = "content.updated",
            Payload = JsonSerializer.Serialize(new
            {
                eventId = Guid.NewGuid().ToString(),
                eventType = "content.updated",
                timestamp = DateTime.UtcNow.ToString("O"),
                source = "catalog-service",
                correlationId = Guid.NewGuid().ToString(),
                data = new
                {
                    contentId = series.Id,
                    contentType = "series",
                    seasonNumber = request.SeasonNumber,
                    episodeNumber = request.EpisodeNumber,
                    updatedAt = series.UpdatedAt.ToString("O")
                }
            }),
            CreatedAt = DateTime.UtcNow
        };

        await _outboxRepository.AddAsync(outboxMessage, cancellationToken);

        return true;
    }
}
