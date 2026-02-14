using System.Text.Json;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Commands.CreateSeries;

public class CreateSeriesHandler : IRequestHandler<CreateSeriesCommand, string>
{
    private readonly ISeriesRepository _seriesRepository;
    private readonly IOutboxRepository _outboxRepository;

    public CreateSeriesHandler(ISeriesRepository seriesRepository, IOutboxRepository outboxRepository)
    {
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _outboxRepository = outboxRepository ?? throw new ArgumentNullException(nameof(outboxRepository));
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

        var outboxMessage = new OutboxMessage
        {
            EventType = "content.created",
            Payload = JsonSerializer.Serialize(new
            {
                eventId = Guid.NewGuid().ToString(),
                eventType = "content.created",
                timestamp = DateTime.UtcNow.ToString("O"),
                source = "catalog-service",
                correlationId = Guid.NewGuid().ToString(),
                data = new
                {
                    contentId = series.Id,
                    contentType = "series",
                    title = series.Title,
                    genres = series.Genres,
                    releaseYear = series.ReleaseYear,
                    creator = series.Creator
                }
            }),
            CreatedAt = DateTime.UtcNow
        };

        await _outboxRepository.AddAsync(outboxMessage, cancellationToken);

        return series.Id;
    }
}
