using System.Text.Json;
using System.Text.RegularExpressions;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using MediatR;
using Microsoft.Extensions.Logging;

namespace CatalogService.Application.Commands.UpdateVideoStatus;

public partial class UpdateVideoStatusHandler : IRequestHandler<UpdateVideoStatusCommand, bool>
{
    private readonly IMovieRepository _movieRepository;
    private readonly ISeriesRepository _seriesRepository;
    private readonly IOutboxRepository _outboxRepository;
    private readonly ILogger<UpdateVideoStatusHandler> _logger;

    public UpdateVideoStatusHandler(
        IMovieRepository movieRepository,
        ISeriesRepository seriesRepository,
        IOutboxRepository outboxRepository,
        ILogger<UpdateVideoStatusHandler> logger)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _outboxRepository = outboxRepository ?? throw new ArgumentNullException(nameof(outboxRepository));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }

    public async Task<bool> Handle(UpdateVideoStatusCommand request, CancellationToken cancellationToken)
    {
        // Try to parse composite content ID for episodes: {seriesId}_s{season}_e{episode}
        var episodeMatch = EpisodeContentIdRegex().Match(request.ContentId);
        if (episodeMatch.Success)
        {
            return await HandleEpisodeAsync(episodeMatch, request, cancellationToken);
        }

        // Otherwise treat as movie
        return await HandleMovieAsync(request, cancellationToken);
    }

    private async Task<bool> HandleMovieAsync(UpdateVideoStatusCommand request, CancellationToken cancellationToken)
    {
        var movie = await _movieRepository.GetByIdAsync(request.ContentId, cancellationToken);

        if (movie is null)
        {
            _logger.LogWarning("Movie with ID {ContentId} not found for video status update", request.ContentId);
            return false;
        }

        if (!IsValidTransition(movie.VideoStatus, request.VideoStatus))
        {
            _logger.LogWarning(
                "Invalid video status transition from {CurrentStatus} to {NewStatus} for movie {ContentId}",
                movie.VideoStatus, request.VideoStatus, request.ContentId);
            return false;
        }

        movie.VideoStatus = request.VideoStatus;
        movie.StreamingInfo = request.StreamingInfo;
        movie.UpdatedAt = DateTime.UtcNow;

        await _movieRepository.UpdateAsync(movie, cancellationToken);

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
                    contentId = movie.Id,
                    contentType = "movie",
                    videoStatus = request.VideoStatus.ToString(),
                    updatedAt = movie.UpdatedAt.ToString("O")
                }
            }),
            CreatedAt = DateTime.UtcNow
        };

        await _outboxRepository.AddAsync(outboxMessage, cancellationToken);

        _logger.LogInformation(
            "Updated video status to {VideoStatus} for movie {ContentId}",
            request.VideoStatus, request.ContentId);

        return true;
    }

    private async Task<bool> HandleEpisodeAsync(
        Match match,
        UpdateVideoStatusCommand request,
        CancellationToken cancellationToken)
    {
        var seriesId = match.Groups["seriesId"].Value;
        var seasonNumber = int.Parse(match.Groups["season"].Value);
        var episodeNumber = int.Parse(match.Groups["episode"].Value);

        var series = await _seriesRepository.GetByIdAsync(seriesId, cancellationToken);
        if (series is null)
        {
            _logger.LogWarning("Series {SeriesId} not found for episode video status update", seriesId);
            return false;
        }

        var season = series.GetSeason(seasonNumber);
        if (season is null)
        {
            _logger.LogWarning("Season {SeasonNumber} not found in series {SeriesId}", seasonNumber, seriesId);
            return false;
        }

        var episode = season.Episodes.FirstOrDefault(e => e.EpisodeNumber == episodeNumber);
        if (episode is null)
        {
            _logger.LogWarning(
                "Episode {EpisodeNumber} not found in season {SeasonNumber} of series {SeriesId}",
                episodeNumber, seasonNumber, seriesId);
            return false;
        }

        if (!IsValidTransition(episode.VideoStatus, request.VideoStatus))
        {
            _logger.LogWarning(
                "Invalid video status transition from {CurrentStatus} to {NewStatus} for episode S{Season}E{Episode} of series {SeriesId}",
                episode.VideoStatus, request.VideoStatus, seasonNumber, episodeNumber, seriesId);
            return false;
        }

        episode.VideoStatus = request.VideoStatus;
        episode.StreamingInfo = request.StreamingInfo;
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
                    contentId = request.ContentId,
                    seriesId,
                    contentType = "episode",
                    seasonNumber,
                    episodeNumber,
                    videoStatus = request.VideoStatus.ToString(),
                    updatedAt = series.UpdatedAt.ToString("O")
                }
            }),
            CreatedAt = DateTime.UtcNow
        };

        await _outboxRepository.AddAsync(outboxMessage, cancellationToken);

        _logger.LogInformation(
            "Updated video status to {VideoStatus} for episode S{Season}E{Episode} of series {SeriesId}",
            request.VideoStatus, seasonNumber, episodeNumber, seriesId);

        return true;
    }

    private static bool IsValidTransition(VideoStatus current, VideoStatus target)
    {
        return (current, target) switch
        {
            (VideoStatus.NotUploaded, VideoStatus.Uploading) => true,
            (VideoStatus.Uploading, VideoStatus.Queued) => true,
            (VideoStatus.Uploading, VideoStatus.Error) => true,
            (VideoStatus.Queued, VideoStatus.Encoding) => true,
            (VideoStatus.Queued, VideoStatus.Error) => true,
            (VideoStatus.Encoding, VideoStatus.Ready) => true,
            (VideoStatus.Encoding, VideoStatus.Error) => true,
            (VideoStatus.Ready, VideoStatus.Uploading) => true,
            (VideoStatus.Error, VideoStatus.Uploading) => true,
            _ => false,
        };
    }

    [GeneratedRegex(@"^(?<seriesId>[a-f0-9]{24})_s(?<season>\d+)_e(?<episode>\d+)$")]
    private static partial Regex EpisodeContentIdRegex();
}
