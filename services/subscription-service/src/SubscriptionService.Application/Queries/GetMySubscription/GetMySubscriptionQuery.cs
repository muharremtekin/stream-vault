using AutoMapper;
using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Application.Queries.GetMySubscription;

public record GetMySubscriptionQuery(Guid UserId) : IRequest<SubscriptionDto?>;

public class GetMySubscriptionHandler : IRequestHandler<GetMySubscriptionQuery, SubscriptionDto?>
{
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IMapper _mapper;

    public GetMySubscriptionHandler(ISubscriptionRepository subscriptionRepository, IMapper mapper)
    {
        _subscriptionRepository = subscriptionRepository;
        _mapper = mapper;
    }

    public async Task<SubscriptionDto?> Handle(GetMySubscriptionQuery request, CancellationToken cancellationToken)
    {
        var subscription = await _subscriptionRepository.GetActiveByUserIdAsync(request.UserId, cancellationToken);
        if (subscription is null)
            return null;

        return _mapper.Map<SubscriptionDto>(subscription);
    }
}
