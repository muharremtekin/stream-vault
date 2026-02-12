using CatalogService.Application.DTOs;
using CatalogService.Application.Queries.GetMovies;
using MediatR;

namespace CatalogService.Application.Queries.GetByGenre;

public record GetByGenreQuery : IRequest<PagedResult<ContentSummaryDto>>
{
    public string Slug { get; init; } = string.Empty;
    public int Page { get; init; } = 1;
    public int PageSize { get; init; } = 20;
}
