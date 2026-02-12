using MediatR;
using UserService.Application.DTOs;
using UserService.Domain.Enums;

namespace UserService.Application.Commands.AddToWatchlist;

public record AddToWatchlistCommand(
    Guid ProfileId,
    string ContentId,
    ContentType ContentType
) : IRequest<WatchlistItemDto>;
