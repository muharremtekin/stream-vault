using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Repositories;

public class InvoiceRepository : IInvoiceRepository
{
    private readonly SubscriptionDbContext _context;

    public InvoiceRepository(SubscriptionDbContext context)
    {
        _context = context;
    }

    public async Task<Invoice?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
    {
        return await _context.Invoices
            .FirstOrDefaultAsync(i => i.Id == id, cancellationToken);
    }

    public async Task AddAsync(Invoice invoice, CancellationToken cancellationToken = default)
    {
        await _context.Invoices.AddAsync(invoice, cancellationToken);
        await _context.SaveChangesAsync(cancellationToken);
    }

    public async Task<List<Invoice>> GetBySubscriptionIdAsync(
        Guid subscriptionId, CancellationToken cancellationToken = default)
    {
        return await _context.Invoices
            .Where(i => i.SubscriptionId == subscriptionId)
            .OrderByDescending(i => i.IssuedAt)
            .ToListAsync(cancellationToken);
    }

    public async Task<string> GenerateInvoiceNumberAsync(CancellationToken cancellationToken = default)
    {
        var today = DateTime.UtcNow.Date;
        var count = await _context.Invoices
            .CountAsync(i => i.IssuedAt >= today, cancellationToken);
        return $"INV-{today:yyyyMMdd}-{(count + 1):D4}";
    }
}
