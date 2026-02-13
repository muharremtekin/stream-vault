using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using MediatR;

namespace CatalogService.Application.Commands.UpdateVideoStatus;

public record UpdateVideoStatusCommand : IRequest<bool>
{
    public string ContentId { get; init; } = string.Empty;
    public VideoStatus VideoStatus { get; init; }
    public StreamingInfo? StreamingInfo { get; init; }
}
