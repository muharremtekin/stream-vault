using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Commands.RefreshToken;

public record RefreshTokenCommand(string RefreshToken) : IRequest<AuthResponseDto>;
