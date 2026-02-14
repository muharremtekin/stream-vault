using System.Text.Json;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using MediatR;
using Microsoft.Extensions.Logging;

namespace CatalogService.Application.Commands.UpdateVideoStatus;

public class UpdateVideoStatusHandler : IRequestHandler<UpdateVideoStatusCommand, bool>
{
    private readonly IMovieRepository _movieRepository;
    private readonly IOutboxRepository _outboxRepository;
    private readonly ILogger<UpdateVideoStatusHandler> _logger;

    public UpdateVideoStatusHandler(
        IMovieRepository movieRepository,
        IOutboxRepository outboxRepository,
        ILogger<UpdateVideoStatusHandler> logger)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _outboxRepository = outboxRepository ?? throw new ArgumentNullException(nameof(outboxRepository));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }

    public async Task<bool> Handle(UpdateVideoStatusCommand request, CancellationToken cancellationToken)
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
}
