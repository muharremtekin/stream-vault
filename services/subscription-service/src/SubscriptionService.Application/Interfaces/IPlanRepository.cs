using SubscriptionService.Application.DTOs;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Application.Interfaces;

public interface IPlanRepository
{
    Task<Plan?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task<List<Plan>> GetAllActiveAsync(CancellationToken cancellationToken = default);
    Task<List<PlanDto>> GetAllActiveDtosAsync(CancellationToken cancellationToken = default);
    Task<Plan?> GetByTierAsync(PlanTier tier, CancellationToken cancellationToken = default);
}
