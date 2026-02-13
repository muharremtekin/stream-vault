using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using UserService.Domain.Entities;
using UserService.Domain.Enums;
using UserService.Infrastructure.Persistence;
using UserService.Infrastructure.Services;
using Xunit;

namespace UserService.UnitTests.Services;

public class JwtTokenServiceTests : IDisposable
{
    private readonly UserDbContext _context;
    private readonly JwtTokenService _tokenService;
    private readonly IConfiguration _configuration;

    public JwtTokenServiceTests()
    {
        var options = new DbContextOptionsBuilder<UserDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;

        _context = new UserDbContext(options);

        var inMemorySettings = new Dictionary<string, string?>
        {
            { "Jwt:Secret", "StreamVault-Test-Secret-Key-Must-Be-At-Least-32-Characters!" },
            { "Jwt:Issuer", "StreamVault.UserService.Test" },
            { "Jwt:Audience", "StreamVault.Client.Test" },
            { "Jwt:ExpiresInMinutes", "60" }
        };

        _configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(inMemorySettings)
            .Build();

        _tokenService = new JwtTokenService(_configuration, _context);
    }

    [Fact]
    public void GenerateAccessToken_ShouldReturnValidJwtToken()
    {
        // Arrange
        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "test@example.com",
            Role = SubscriptionTier.Premium
        };

        // Act
        var token = _tokenService.GenerateAccessToken(user);

        // Assert
        Assert.NotNull(token);
        Assert.NotEmpty(token);

        var handler = new JwtSecurityTokenHandler();
        var jwtToken = handler.ReadJwtToken(token);

        Assert.Equal("StreamVault.UserService.Test", jwtToken.Issuer);
        Assert.Contains(jwtToken.Claims, c => c.Type == JwtRegisteredClaimNames.Email && c.Value == user.Email);
        Assert.Contains(jwtToken.Claims, c => c.Type == JwtRegisteredClaimNames.Sub && c.Value == user.Id.ToString());
    }

    [Fact]
    public void GenerateAccessToken_ShouldIncludeRoleClaim()
    {
        // Arrange
        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "admin@example.com",
            Role = SubscriptionTier.Premium
        };

        // Act
        var token = _tokenService.GenerateAccessToken(user);

        // Assert
        var handler = new JwtSecurityTokenHandler();
        var jwtToken = handler.ReadJwtToken(token);

        // JwtTokenService uses new Claim("role", ...) directly, not ClaimTypes.Role
        Assert.Contains(jwtToken.Claims, c =>
            c.Type == "role" && c.Value == "Premium");
    }

    [Fact]
    public void GenerateRefreshToken_ShouldReturnNonEmptyToken()
    {
        // Arrange
        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "test@example.com",
            Role = SubscriptionTier.Free
        };

        // Act
        var refreshToken = _tokenService.GenerateRefreshToken(user);

        // Assert
        Assert.NotNull(refreshToken);
        Assert.NotEmpty(refreshToken);
    }

    [Fact]
    public void GenerateRefreshToken_ShouldPersistTokenInDatabase()
    {
        // Arrange
        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "test@example.com",
            Role = SubscriptionTier.Free
        };

        _context.Users.Add(user);
        _context.SaveChanges();

        // Act
        var refreshToken = _tokenService.GenerateRefreshToken(user);

        // Assert
        var storedToken = _context.RefreshTokens.FirstOrDefault(rt => rt.Token == refreshToken);
        Assert.NotNull(storedToken);
        Assert.Equal(user.Id, storedToken.UserId);
        Assert.False(storedToken.IsRevoked);
    }

    [Fact]
    public void GenerateAccessToken_ShouldBeValidatable_WithCorrectKey()
    {
        // Arrange
        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "validate@example.com",
            Role = SubscriptionTier.Premium
        };

        // Act
        var token = _tokenService.GenerateAccessToken(user);

        // Assert — validate the token programmatically
        var tokenHandler = new JwtSecurityTokenHandler();
        var key = System.Text.Encoding.UTF8.GetBytes("StreamVault-Test-Secret-Key-Must-Be-At-Least-32-Characters!");
        var validationParams = new Microsoft.IdentityModel.Tokens.TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidIssuer = "StreamVault.UserService.Test",
            ValidateAudience = true,
            ValidAudience = "StreamVault.Client.Test",
            ValidateLifetime = true,
            IssuerSigningKey = new Microsoft.IdentityModel.Tokens.SymmetricSecurityKey(key),
            ClockSkew = TimeSpan.Zero
        };

        var principal = tokenHandler.ValidateToken(token, validationParams, out var validatedToken);
        Assert.NotNull(principal);
        Assert.NotNull(validatedToken);
        // .NET maps "sub" → ClaimTypes.NameIdentifier when validating
        Assert.Equal(user.Id.ToString(),
            principal.FindFirst(ClaimTypes.NameIdentifier)?.Value);
        Assert.Equal(user.Email,
            principal.FindFirst(ClaimTypes.Email)?.Value);
    }

    [Fact]
    public void GenerateAccessToken_WithExpiredToken_ShouldFailValidation()
    {
        // Arrange — service with 0-minute expiry
        var expiredSettings = new Dictionary<string, string?>
        {
            { "Jwt:Secret", "StreamVault-Test-Secret-Key-Must-Be-At-Least-32-Characters!" },
            { "Jwt:Issuer", "StreamVault.UserService.Test" },
            { "Jwt:Audience", "StreamVault.Client.Test" },
            { "Jwt:ExpiresInMinutes", "0" }
        };
        var expiredConfig = new ConfigurationBuilder()
            .AddInMemoryCollection(expiredSettings)
            .Build();
        var expiredService = new JwtTokenService(expiredConfig, _context);

        var user = new User
        {
            Id = Guid.NewGuid(),
            Email = "expired@example.com",
            Role = SubscriptionTier.Free
        };

        // Act
        var token = expiredService.GenerateAccessToken(user);

        // Assert — validation should fail due to expiry
        var tokenHandler = new JwtSecurityTokenHandler();
        var key = System.Text.Encoding.UTF8.GetBytes("StreamVault-Test-Secret-Key-Must-Be-At-Least-32-Characters!");
        var validationParams = new Microsoft.IdentityModel.Tokens.TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidIssuer = "StreamVault.UserService.Test",
            ValidateAudience = true,
            ValidAudience = "StreamVault.Client.Test",
            ValidateLifetime = true,
            IssuerSigningKey = new Microsoft.IdentityModel.Tokens.SymmetricSecurityKey(key),
            ClockSkew = TimeSpan.Zero
        };

        Assert.Throws<Microsoft.IdentityModel.Tokens.SecurityTokenExpiredException>(() =>
            tokenHandler.ValidateToken(token, validationParams, out _));
    }

    public void Dispose()
    {
        _context.Dispose();
    }
}
