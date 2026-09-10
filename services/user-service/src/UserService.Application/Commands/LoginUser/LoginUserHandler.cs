using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Exceptions;

namespace UserService.Application.Commands.LoginUser;

public class LoginUserHandler : IRequestHandler<LoginUserCommand, AuthResponseDto>
{
    private readonly IUserRepository _userRepository;
    private readonly IPasswordHasher _passwordHasher;
    private readonly ITokenService _tokenService;

    public LoginUserHandler(
        IUserRepository userRepository,
        IPasswordHasher passwordHasher,
        ITokenService tokenService)
    {
        _userRepository = userRepository;
        _passwordHasher = passwordHasher;
        _tokenService = tokenService;
    }

    public async Task<AuthResponseDto> Handle(LoginUserCommand request, CancellationToken cancellationToken)
    {
        var user = await _userRepository.GetSummaryByEmailAsync(request.Email, cancellationToken);
        if (user is null)
        {
            throw new UserNotFoundException($"Invalid email or password.");
        }

        var isValidPassword = _passwordHasher.Verify(request.Password, user.PasswordHash);
        if (!isValidPassword)
        {
            throw new UserNotFoundException("Invalid email or password.");
        }

        var tokenUser = new User
        {
            Id = user.Id,
            Email = user.Email,
            PasswordHash = user.PasswordHash,
            Role = user.Role
        };

        var accessToken = _tokenService.GenerateAccessToken(tokenUser);
        var refreshToken = await _tokenService.GenerateRefreshTokenAsync(tokenUser, cancellationToken);

        return new AuthResponseDto
        {
            AccessToken = accessToken,
            RefreshToken = refreshToken,
            ExpiresIn = _tokenService.AccessTokenExpirationSeconds,
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
