namespace SubscriptionService.Application.DTOs;

public class InvoiceDto
{
    public Guid Id { get; set; }
    public Guid SubscriptionId { get; set; }
    public string InvoiceNumber { get; set; } = string.Empty;
    public decimal Amount { get; set; }
    public string Currency { get; set; } = string.Empty;
    public DateTime PeriodStart { get; set; }
    public DateTime PeriodEnd { get; set; }
    public DateTime IssuedAt { get; set; }
}
