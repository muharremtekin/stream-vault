using CatalogService.Application.DTOs;
using MediatR;

namespace CatalogService.Application.Queries.GetStreamingInfo;

public record GetStreamingInfoQuery : IRequest<StreamingInfoDto?>
{
    public string MovieId { get; init; } = string.Empty;
}
