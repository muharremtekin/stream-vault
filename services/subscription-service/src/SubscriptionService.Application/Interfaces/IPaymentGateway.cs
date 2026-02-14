namespace SubscriptionService.Application.Interfaces;

public interface IPaymentGateway
{
    Task<PaymentResult> ChargeAsync(string cardNumber, decimal amount, string currency, CancellationToken cancellationToken = default);
    Task<PaymentResult> RefundAsync(string transactionId, decimal amount, CancellationToken cancellationToken = default);
}

public record PaymentResult(
    bool IsSuccess,
    string? TransactionId,
    string? FailureReason
);
