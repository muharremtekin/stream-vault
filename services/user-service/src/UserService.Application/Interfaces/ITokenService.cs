using UserService.Domain.Entities;

namespace UserService.Application.Interfaces;

public interface ITokenService
{
    string GenerateAccessToken(User user);

    string GenerateRefreshToken(User user);

    Task<bool> ValidateRefreshToken(string token, CancellationToken cancellationToken = default);

    Task<Guid?> GetUserIdFromRefreshToken(string token, CancellationToken cancellationToken = default);

    Task RevokeRefreshToken(string token, CancellationToken cancellationToken = default);
}
