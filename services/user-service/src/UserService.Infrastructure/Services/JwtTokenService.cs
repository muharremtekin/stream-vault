using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Security.Cryptography;
using System.Text;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.IdentityModel.Tokens;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Infrastructure.Persistence;

namespace UserService.Infrastructure.Services;

public class JwtTokenService : ITokenService
{
    private readonly IConfiguration _configuration;
    private readonly UserDbContext _context;

    public JwtTokenService(IConfiguration configuration, UserDbContext context)
    {
        _configuration = configuration;
        _context = context;
    }

    public string GenerateAccessToken(User user)
    {
        var key = new SymmetricSecurityKey(
            Encoding.UTF8.GetBytes(_configuration["Jwt:Secret"]
                ?? throw new InvalidOperationException("JWT secret is not configured.")));

        var credentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var claims = new List<Claim>
        {
            new(JwtRegisteredClaimNames.Sub, user.Id.ToString()),
            new(JwtRegisteredClaimNames.Email, user.Email),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
            new(ClaimTypes.Role, user.Role.ToString())
        };

        var expiresInMinutes = int.Parse(
            _configuration["Jwt:ExpiresInMinutes"] ?? "60");

        var token = new JwtSecurityToken(
            issuer: _configuration["Jwt:Issuer"],
            audience: _configuration["Jwt:Audience"],
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(expiresInMinutes),
            signingCredentials: credentials);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    public string GenerateRefreshToken(User user)
    {
        var randomBytes = new byte[64];
        using var rng = RandomNumberGenerator.Create();
        rng.GetBytes(randomBytes);

        var refreshTokenString = Convert.ToBase64String(randomBytes);

        var refreshToken = new RefreshToken
        {
            Id = Guid.NewGuid(),
            UserId = user.Id,
            Token = refreshTokenString,
            ExpiresAt = DateTime.UtcNow.AddDays(7),
            CreatedAt = DateTime.UtcNow,
            IsRevoked = false
        };

        _context.RefreshTokens.Add(refreshToken);
        _context.SaveChanges();

        return refreshTokenString;
    }

    public async Task<bool> ValidateRefreshToken(string token, CancellationToken cancellationToken = default)
    {
        var refreshToken = await _context.RefreshTokens
            .FirstOrDefaultAsync(rt => rt.Token == token, cancellationToken);

        if (refreshToken is null)
        {
            return false;
        }

        return !refreshToken.IsRevoked && refreshToken.ExpiresAt > DateTime.UtcNow;
    }

    public async Task<Guid?> GetUserIdFromRefreshToken(string token, CancellationToken cancellationToken = default)
    {
        var refreshToken = await _context.RefreshTokens
            .FirstOrDefaultAsync(rt => rt.Token == token, cancellationToken);

        if (refreshToken is null || refreshToken.IsRevoked || refreshToken.ExpiresAt <= DateTime.UtcNow)
        {
            return null;
        }

        return refreshToken.UserId;
    }

    public async Task RevokeRefreshToken(string token, CancellationToken cancellationToken = default)
    {
        var refreshToken = await _context.RefreshTokens
            .FirstOrDefaultAsync(rt => rt.Token == token, cancellationToken);

        if (refreshToken is not null)
        {
            refreshToken.IsRevoked = true;
            await _context.SaveChangesAsync(cancellationToken);
        }
    }
}
