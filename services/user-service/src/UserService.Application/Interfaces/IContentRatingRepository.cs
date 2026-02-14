using UserService.Domain.Entities;

namespace UserService.Application.Interfaces;

public interface IContentRatingRepository
{
    Task<ContentRating?> GetAsync(Guid userId, string contentId, CancellationToken cancellationToken = default);
    Task<List<ContentRating>> GetByUserIdAsync(Guid userId, int page, int pageSize, CancellationToken cancellationToken = default);
    Task UpsertAsync(ContentRating rating, CancellationToken cancellationToken = default);
    Task<int> CountByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
}
