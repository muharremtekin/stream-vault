using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Queries.GetWatchlist;

public record GetWatchlistQuery(
    Guid ProfileId
) : IRequest<List<WatchlistItemDto>>;
