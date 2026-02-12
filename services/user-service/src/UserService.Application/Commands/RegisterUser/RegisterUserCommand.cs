using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Commands.RegisterUser;

public record RegisterUserCommand(
    string Email,
    string Password
) : IRequest<AuthResponseDto>;
