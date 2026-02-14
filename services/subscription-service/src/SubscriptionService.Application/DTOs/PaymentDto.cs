namespace SubscriptionService.Application.DTOs;

public class PaymentDto
{
    public Guid Id { get; set; }
    public Guid SubscriptionId { get; set; }
    public decimal Amount { get; set; }
    public string Currency { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty;
    public string? TransactionId { get; set; }
    public string CardLastFour { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; }
}
