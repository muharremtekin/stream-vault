using UserService.Domain.Entities;

namespace UserService.Application.Interfaces;

public interface ITokenService
{
    string GenerateAccessToken(User user);

    Task<string> GenerateRefreshTokenAsync(User user, CancellationToken cancellationToken = default);

    Task<Guid?> ConsumeRefreshTokenAsync(string token, CancellationToken cancellationToken = default);
}
