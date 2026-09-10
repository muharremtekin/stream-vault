using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Application.Queries.GetPlans;

public record GetPlansQuery : IRequest<List<PlanDto>>;

public class GetPlansHandler : IRequestHandler<GetPlansQuery, List<PlanDto>>
{
    private readonly IPlanRepository _planRepository;

    public GetPlansHandler(IPlanRepository planRepository)
    {
        _planRepository = planRepository;
    }

    public async Task<List<PlanDto>> Handle(GetPlansQuery request, CancellationToken cancellationToken)
    {
        return await _planRepository.GetAllActiveDtosAsync(cancellationToken);
    }
}
