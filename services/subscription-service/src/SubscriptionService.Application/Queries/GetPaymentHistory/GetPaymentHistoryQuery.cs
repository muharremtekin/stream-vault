using AutoMapper;
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
    private readonly IMapper _mapper;

    public GetPaymentHistoryHandler(IPaymentRepository paymentRepository, IMapper mapper)
    {
        _paymentRepository = paymentRepository;
        _mapper = mapper;
    }

    public async Task<List<PaymentDto>> Handle(GetPaymentHistoryQuery request, CancellationToken cancellationToken)
    {
        var payments = await _paymentRepository.GetByUserIdAsync(
            request.UserId, request.Limit, request.Offset, cancellationToken);
        return _mapper.Map<List<PaymentDto>>(payments);
    }
}
