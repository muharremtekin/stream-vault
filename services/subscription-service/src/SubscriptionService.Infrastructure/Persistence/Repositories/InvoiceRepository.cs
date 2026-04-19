using Microsoft.EntityFrameworkCore;
using SubscriptionService.Application.DTOs;
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
    }

    public async Task<List<Invoice>> GetBySubscriptionIdAsync(
        Guid subscriptionId, CancellationToken cancellationToken = default)
    {
        return await _context.Invoices
            .AsNoTracking()
            .Where(i => i.SubscriptionId == subscriptionId)
            .OrderByDescending(i => i.IssuedAt)
            .ToListAsync(cancellationToken);
    }

    public async Task<List<InvoiceDto>> GetDtosBySubscriptionIdAsync(
        Guid subscriptionId,
        CancellationToken cancellationToken = default)
    {
        return await _context.Invoices
            .AsNoTracking()
            .Where(i => i.SubscriptionId == subscriptionId)
            .OrderByDescending(i => i.IssuedAt)
            .Select(i => new InvoiceDto
            {
                Id = i.Id,
                SubscriptionId = i.SubscriptionId,
                InvoiceNumber = i.InvoiceNumber,
                Amount = i.Amount,
                Currency = i.Currency,
                PeriodStart = i.PeriodStart,
                PeriodEnd = i.PeriodEnd,
                IssuedAt = i.IssuedAt
            })
            .ToListAsync(cancellationToken);
    }

    public async Task<string> GenerateInvoiceNumberAsync(CancellationToken cancellationToken = default)
    {
        var today = DateTime.UtcNow.Date;

        if (_context.Database.IsRelational())
        {
            var nextValue = await _context.Database
                .SqlQuery<long>($"SELECT nextval('invoice_numbers')")
                .SingleAsync(cancellationToken);

            return $"INV-{today:yyyyMMdd}-{nextValue:D6}";
        }

        var nextValueFallback = await _context.Invoices.CountAsync(cancellationToken) + 1;
        return $"INV-{today:yyyyMMdd}-{nextValueFallback:D6}";
    }
}
