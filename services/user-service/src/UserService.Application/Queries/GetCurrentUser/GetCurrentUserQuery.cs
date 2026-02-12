using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Queries.GetCurrentUser;

public record GetCurrentUserQuery(
    Guid UserId
) : IRequest<UserDto>;
