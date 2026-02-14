using Microsoft.Extensions.Logging;
using SubscriptionService.Application.Interfaces;

namespace SubscriptionService.Infrastructure.Services;

public class MockPaymentGateway : IPaymentGateway
{
    private readonly ILogger<MockPaymentGateway> _logger;

    public MockPaymentGateway(ILogger<MockPaymentGateway> logger)
    {
        _logger = logger;
    }

    public async Task<PaymentResult> ChargeAsync(
        string cardNumber,
        decimal amount,
        string currency,
        CancellationToken cancellationToken = default)
    {
        var delay = Random.Shared.Next(200, 801);
        await Task.Delay(delay, cancellationToken);

        var sanitized = cardNumber.Replace(" ", "").Replace("-", "");

        _logger.LogInformation(
            "Processing charge of {Amount} {Currency} for card ending {Last4}. Delay: {Delay}ms",
            amount, currency, sanitized[^4..], delay);

        return sanitized switch
        {
            "4242424242424242" => new PaymentResult(
                true,
                $"txn_{Guid.NewGuid():N}",
                null),
            "4000000000000002" => new PaymentResult(
                false,
                null,
                "Insufficient funds"),
            "4000000000000069" => new PaymentResult(
                false,
                null,
                "Expired card"),
            "4000000000000127" => new PaymentResult(
                false,
                null,
                "General payment error"),
            _ => new PaymentResult(
                true,
                $"txn_{Guid.NewGuid():N}",
                null)
        };
    }

    public async Task<PaymentResult> RefundAsync(
        string transactionId,
        decimal amount,
        CancellationToken cancellationToken = default)
    {
        var delay = Random.Shared.Next(200, 801);
        await Task.Delay(delay, cancellationToken);

        _logger.LogInformation(
            "Processing refund of {Amount} for transaction {TransactionId}. Delay: {Delay}ms",
            amount, transactionId, delay);

        return new PaymentResult(
            true,
            $"ref_{Guid.NewGuid():N}",
            null);
    }
}
