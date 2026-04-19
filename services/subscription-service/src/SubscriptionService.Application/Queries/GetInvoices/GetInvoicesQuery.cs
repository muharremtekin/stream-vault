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

    public GetInvoicesHandler(
        ISubscriptionRepository subscriptionRepository,
        IInvoiceRepository invoiceRepository)
    {
        _subscriptionRepository = subscriptionRepository;
        _invoiceRepository = invoiceRepository;
    }

    public async Task<List<InvoiceDto>> Handle(GetInvoicesQuery request, CancellationToken cancellationToken)
    {
        var subscriptionId = await _subscriptionRepository.GetActiveSubscriptionIdByUserIdAsync(
            request.UserId,
            cancellationToken);
        if (!subscriptionId.HasValue)
            throw new SubscriptionNotFoundException(request.UserId);

        return await _invoiceRepository.GetDtosBySubscriptionIdAsync(
            subscriptionId.Value,
            cancellationToken);
    }
}
