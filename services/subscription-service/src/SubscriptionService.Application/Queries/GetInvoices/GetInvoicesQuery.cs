using AutoMapper;
using MediatR;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.Application.Queries.GetInvoices;

public record GetInvoicesQuery(Guid UserId) : IRequest<List<InvoiceDto>>;

public class GetInvoicesHandler : IRequestHandler<GetInvoicesQuery, List<InvoiceDto>>
{
    private readonly ISubscriptionRepository _subscriptionRepository;
    private readonly IInvoiceRepository _invoiceRepository;
    private readonly IMapper _mapper;

    public GetInvoicesHandler(
        ISubscriptionRepository subscriptionRepository,
        IInvoiceRepository invoiceRepository,
        IMapper mapper)
    {
        _subscriptionRepository = subscriptionRepository;
        _invoiceRepository = invoiceRepository;
        _mapper = mapper;
    }

    public async Task<List<InvoiceDto>> Handle(GetInvoicesQuery request, CancellationToken cancellationToken)
    {
        var subscription = await _subscriptionRepository.GetActiveByUserIdAsync(request.UserId, cancellationToken);
        if (subscription is null)
            throw new SubscriptionNotFoundException(request.UserId);

        var invoices = await _invoiceRepository.GetBySubscriptionIdAsync(subscription.Id, cancellationToken);
        return _mapper.Map<List<InvoiceDto>>(invoices);
    }
}
