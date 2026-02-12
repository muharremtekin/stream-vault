using CatalogService.Application.DTOs;
using MediatR;

namespace CatalogService.Application.Queries.GetMovies;

public record GetMoviesQuery : IRequest<PagedResult<MovieDto>>
{
    public int Page { get; init; } = 1;
    public int PageSize { get; init; } = 20;
    public string? Genre { get; init; }
    public string? Sort { get; init; }
    public int? Year { get; init; }
}

public record PagedResult<T>
{
    public List<T> Items { get; init; } = new();
    public int TotalCount { get; init; }
    public int Page { get; init; }
    public int PageSize { get; init; }
    public int TotalPages => (int)Math.Ceiling(TotalCount / (double)PageSize);
    public bool HasNextPage => Page < TotalPages;
    public bool HasPreviousPage => Page > 1;
}
