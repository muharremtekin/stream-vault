using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Application.Queries.GetMySubscription;

public record GetMySubscriptionQuery(Guid UserId) : IRequest<SubscriptionDto?>;

public class GetMySubscriptionHandler : IRequestHandler<GetMySubscriptionQuery, SubscriptionDto?>
{
    private readonly ISubscriptionRepository _subscriptionRepository;

    public GetMySubscriptionHandler(ISubscriptionRepository subscriptionRepository)
    {
        _subscriptionRepository = subscriptionRepository;
    }

    public async Task<SubscriptionDto?> Handle(GetMySubscriptionQuery request, CancellationToken cancellationToken)
    {
        return await _subscriptionRepository.GetActiveDtoByUserIdAsync(
            request.UserId,
            cancellationToken);
    }
}
