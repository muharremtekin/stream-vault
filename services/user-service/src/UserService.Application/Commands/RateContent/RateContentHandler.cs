using System.Text.Json;
using MediatR;
using Microsoft.Extensions.Logging;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;

namespace UserService.Application.Commands.RateContent;

public class RateContentHandler : IRequestHandler<RateContentCommand, ContentRatingDto>
{
    private readonly IContentRatingRepository _ratingRepository;
    private readonly IRabbitMqPublisher _publisher;
    private readonly ILogger<RateContentHandler> _logger;

    public RateContentHandler(
        IContentRatingRepository ratingRepository,
        IRabbitMqPublisher publisher,
        ILogger<RateContentHandler> logger)
    {
        _ratingRepository = ratingRepository;
        _publisher = publisher;
        _logger = logger;
    }

    public async Task<ContentRatingDto> Handle(RateContentCommand request, CancellationToken cancellationToken)
    {
        var rating = new ContentRating
        {
            UserId = request.UserId,
            ContentId = request.ContentId,
            Rating = request.Rating,
            RatedAt = DateTime.UtcNow
        };

        await _ratingRepository.UpsertAsync(rating, cancellationToken);

        try
        {
            var payload = JsonSerializer.Serialize(new
            {
                eventId = Guid.NewGuid().ToString(),
                eventType = "content.rated",
                timestamp = DateTime.UtcNow.ToString("O"),
                source = "user-service",
                correlationId = Guid.NewGuid().ToString(),
                data = new
                {
                    userId = request.UserId.ToString(),
                    contentId = request.ContentId,
                    rating = request.Rating,
                    ratedAt = rating.RatedAt.ToString("O")
                }
            });

            await _publisher.PublishAsync("user.events", "content.rated", payload, cancellationToken);
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Failed to publish content.rated event for user {UserId}, content {ContentId}",
                request.UserId, request.ContentId);
        }

        return new ContentRatingDto(rating.Id, rating.ContentId, rating.Rating, rating.RatedAt);
    }
}
