using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Enums;

namespace UserService.Infrastructure.Persistence;

public class DbSeeder
{
    private readonly UserDbContext _context;
    private readonly IPasswordHasher _passwordHasher;
    private readonly ILogger<DbSeeder> _logger;

    private static readonly (string Email, string Password, SubscriptionTier Role)[] SeedUsers =
    [
        ("admin@streamvault.com",    "SeedPass123@", SubscriptionTier.Admin),
        ("premium@streamvault.com",  "SeedPass123@", SubscriptionTier.Premium),
        ("standard@streamvault.com", "SeedPass123@", SubscriptionTier.Standard),
        ("basic@streamvault.com",    "SeedPass123@", SubscriptionTier.Basic),
        ("free@streamvault.com",     "SeedPass123@", SubscriptionTier.Free),
    ];

    public DbSeeder(
        UserDbContext context,
        IPasswordHasher passwordHasher,
        ILogger<DbSeeder> logger)
    {
        _context = context;
        _passwordHasher = passwordHasher;
        _logger = logger;
    }

    public async Task SeedAsync(CancellationToken cancellationToken = default)
    {
        foreach (var (email, password, role) in SeedUsers)
        {
            var exists = await _context.Users.AnyAsync(u => u.Email == email, cancellationToken);
            if (exists)
            {
                _logger.LogDebug("Seed user {Email} already exists. Skipping.", email);
                continue;
            }

            var user = new User
            {
                Id = Guid.NewGuid(),
                Email = email,
                PasswordHash = _passwordHasher.Hash(password),
                Role = role,
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow
            };

            _context.Users.Add(user);
            _logger.LogInformation("Seeded user {Email} with role {Role}.", email, role);
        }

        await _context.SaveChangesAsync(cancellationToken);
    }
}
