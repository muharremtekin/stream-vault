using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Exceptions;

namespace UserService.Application.Commands.RefreshToken;

public class RefreshTokenHandler : IRequestHandler<RefreshTokenCommand, AuthResponseDto>
{
    private readonly ITokenService _tokenService;
    private readonly IUserRepository _userRepository;

    public RefreshTokenHandler(ITokenService tokenService, IUserRepository userRepository)
    {
        _tokenService = tokenService;
        _userRepository = userRepository;
    }

    public async Task<AuthResponseDto> Handle(RefreshTokenCommand request, CancellationToken cancellationToken)
    {
        var userId = await _tokenService.ConsumeRefreshTokenAsync(request.RefreshToken, cancellationToken);
        if (userId is null)
        {
            throw new InvalidOperationException("Invalid or expired refresh token.");
        }

        var user = await _userRepository.GetSummaryByIdAsync(userId.Value, cancellationToken);
        if (user is null)
        {
            throw new UserNotFoundException("User associated with refresh token not found.");
        }

        var tokenUser = new User
        {
            Id = user.Id,
            Email = user.Email,
            PasswordHash = user.PasswordHash,
            Role = user.Role
        };

        var accessToken = _tokenService.GenerateAccessToken(tokenUser);
        var newRefreshToken = await _tokenService.GenerateRefreshTokenAsync(tokenUser, cancellationToken);

        return new AuthResponseDto
        {
            AccessToken = accessToken,
            RefreshToken = newRefreshToken,
            ExpiresIn = 3600,
            User = new UserDto
            {
                Id = user.Id,
                Email = user.Email,
                Role = user.Role.ToString(),
                ProfileCount = user.ProfileCount
            }
        };
    }
}
