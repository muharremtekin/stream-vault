using CatalogService.Application.DTOs;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Queries.GetContentById;

public record GetContentByIdQuery : IRequest<ContentSummaryDto?>
{
    public string Id { get; init; } = string.Empty;
    public ContentType ContentType { get; init; }
}
