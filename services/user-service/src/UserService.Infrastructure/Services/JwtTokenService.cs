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
    private readonly UserDbContext _context;
    private readonly JwtSecurityTokenHandler _tokenHandler;
    private readonly SigningCredentials _signingCredentials;
    private readonly string? _issuer;
    private readonly string? _audience;
    private readonly int _expiresInMinutes;
    private readonly int _refreshTokenExpirationDays;

    public JwtTokenService(IConfiguration configuration, UserDbContext context)
    {
        _context = context;
        _tokenHandler = new JwtSecurityTokenHandler();

        var secret = configuration["Jwt:Secret"]
            ?? throw new InvalidOperationException("JWT secret is not configured.");

        _signingCredentials = new SigningCredentials(
            new SymmetricSecurityKey(Encoding.UTF8.GetBytes(secret)),
            SecurityAlgorithms.HmacSha256);

        _issuer = configuration["Jwt:Issuer"];
        _audience = configuration["Jwt:Audience"];
        _expiresInMinutes = int.Parse(configuration["Jwt:ExpiresInMinutes"] ?? "60");
        _refreshTokenExpirationDays = int.Parse(configuration["Jwt:RefreshTokenExpirationDays"] ?? "30");
    }

    public string GenerateAccessToken(User user)
    {
        var claims = new List<Claim>
        {
            new(JwtRegisteredClaimNames.Sub, user.Id.ToString()),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
            new("role", user.Role.ToString())
        };

        var token = new JwtSecurityToken(
            issuer: _issuer,
            audience: _audience,
            claims: claims,
            expires: DateTime.UtcNow.AddMinutes(_expiresInMinutes),
            signingCredentials: _signingCredentials);

        return _tokenHandler.WriteToken(token);
    }

    public async Task<string> GenerateRefreshTokenAsync(User user, CancellationToken cancellationToken = default)
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
            ExpiresAt = DateTime.UtcNow.AddDays(_refreshTokenExpirationDays),
            CreatedAt = DateTime.UtcNow,
            IsRevoked = false
        };

        await _context.RefreshTokens.AddAsync(refreshToken, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);

        return refreshTokenString;
    }

    public async Task<Guid?> ConsumeRefreshTokenAsync(string token, CancellationToken cancellationToken = default)
    {
        var now = DateTime.UtcNow;
        var refreshToken = await _context.RefreshTokens
            .AsNoTracking()
            .Select(rt => new
            {
                rt.Id,
                rt.UserId,
                rt.Token,
                rt.IsRevoked,
                rt.ExpiresAt
            })
            .FirstOrDefaultAsync(
                rt => rt.Token == token && !rt.IsRevoked && rt.ExpiresAt > now,
                cancellationToken);

        if (refreshToken is null)
        {
            return null;
        }

        if (_context.Database.IsRelational())
        {
            var updatedRows = await _context.RefreshTokens
                .Where(rt => rt.Id == refreshToken.Id && !rt.IsRevoked && rt.ExpiresAt > now)
                .ExecuteUpdateAsync(
                    setters => setters.SetProperty(rt => rt.IsRevoked, true),
                    cancellationToken);

            return updatedRows == 1 ? refreshToken.UserId : null;
        }

        var trackedToken = await _context.RefreshTokens
            .FirstOrDefaultAsync(
                rt => rt.Id == refreshToken.Id && !rt.IsRevoked && rt.ExpiresAt > now,
                cancellationToken);

        if (trackedToken is null)
        {
            return null;
        }

        trackedToken.IsRevoked = true;
        await _context.SaveChangesAsync(cancellationToken);
        return trackedToken.UserId;
    }
}
