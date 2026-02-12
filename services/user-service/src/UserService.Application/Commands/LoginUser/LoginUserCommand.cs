using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Commands.LoginUser;

public record LoginUserCommand(
    string Email,
    string Password
) : IRequest<AuthResponseDto>;
