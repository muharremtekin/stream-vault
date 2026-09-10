using UserService.Domain.Entities;

namespace UserService.Application.Interfaces;

public interface IProfileRepository
{
    Task<List<Profile>> GetByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);

    Task<Profile?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);

    Task<bool> ExistsAsync(Guid id, CancellationToken cancellationToken = default);

    Task AddAsync(Profile profile, CancellationToken cancellationToken = default);

    Task<int> CountByUserIdAsync(Guid userId, CancellationToken cancellationToken = default);
}
