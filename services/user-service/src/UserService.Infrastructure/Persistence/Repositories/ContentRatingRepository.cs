using Microsoft.EntityFrameworkCore;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Repositories;

public class ContentRatingRepository : IContentRatingRepository
{
    private readonly UserDbContext _context;

    public ContentRatingRepository(UserDbContext context)
    {
        _context = context;
    }

    public async Task<ContentRating?> GetAsync(Guid userId, string contentId, CancellationToken cancellationToken = default)
    {
        return await _context.ContentRatings
            .AsNoTracking()
            .FirstOrDefaultAsync(r => r.UserId == userId && r.ContentId == contentId, cancellationToken);
    }

    public async Task<List<ContentRating>> GetByUserIdAsync(Guid userId, int page, int pageSize, CancellationToken cancellationToken = default)
    {
        return await _context.ContentRatings
            .AsNoTracking()
            .Where(r => r.UserId == userId)
            .OrderByDescending(r => r.RatedAt)
            .Skip((page - 1) * pageSize)
            .Take(pageSize)
            .ToListAsync(cancellationToken);
    }

    public async Task UpsertAsync(ContentRating rating, CancellationToken cancellationToken = default)
    {
        var existing = await _context.ContentRatings
            .FirstOrDefaultAsync(r => r.UserId == rating.UserId && r.ContentId == rating.ContentId, cancellationToken);

        if (existing is not null)
        {
            rating.Id = existing.Id;
            existing.Rating = rating.Rating;
            existing.RatedAt = rating.RatedAt;
        }
        else
        {
            await _context.ContentRatings.AddAsync(rating, cancellationToken);
        }

        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task<int> CountByUserIdAsync(Guid userId, CancellationToken cancellationToken = default)
    {
        return await _context.ContentRatings
            .AsNoTracking()
            .CountAsync(r => r.UserId == userId, cancellationToken);
    }
}
