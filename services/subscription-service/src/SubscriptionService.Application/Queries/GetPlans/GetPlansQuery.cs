using AutoMapper;
using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Application.Queries.GetPlans;

public record GetPlansQuery : IRequest<List<PlanDto>>;

public class GetPlansHandler : IRequestHandler<GetPlansQuery, List<PlanDto>>
{
    private readonly IPlanRepository _planRepository;
    private readonly IMapper _mapper;

    public GetPlansHandler(IPlanRepository planRepository, IMapper mapper)
    {
        _planRepository = planRepository;
        _mapper = mapper;
    }

    public async Task<List<PlanDto>> Handle(GetPlansQuery request, CancellationToken cancellationToken)
    {
        var plans = await _planRepository.GetAllActiveAsync(cancellationToken);
        return _mapper.Map<List<PlanDto>>(plans);
    }
}
