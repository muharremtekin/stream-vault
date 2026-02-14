using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Interfaces;

public interface IInvoiceRepository
{
    Task<Invoice?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default);
    Task AddAsync(Invoice invoice, CancellationToken cancellationToken = default);
    Task<List<Invoice>> GetBySubscriptionIdAsync(Guid subscriptionId, CancellationToken cancellationToken = default);
    Task<string> GenerateInvoiceNumberAsync(CancellationToken cancellationToken = default);
}
