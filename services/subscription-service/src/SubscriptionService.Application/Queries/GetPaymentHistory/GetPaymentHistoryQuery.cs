using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Application.Queries.GetPaymentHistory;

public record GetPaymentHistoryQuery(
    Guid UserId,
    int Limit = 20,
    int Offset = 0
) : IRequest<List<PaymentDto>>;

public class GetPaymentHistoryHandler : IRequestHandler<GetPaymentHistoryQuery, List<PaymentDto>>
{
    private readonly IPaymentRepository _paymentRepository;

    public GetPaymentHistoryHandler(IPaymentRepository paymentRepository)
    {
        _paymentRepository = paymentRepository;
    }

    public async Task<List<PaymentDto>> Handle(GetPaymentHistoryQuery request, CancellationToken cancellationToken)
    {
        return await _paymentRepository.GetHistoryByUserIdAsync(
            request.UserId,
            request.Limit,
            request.Offset,
            cancellationToken);
    }
}
