using FluentAssertions;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Application.Sagas;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.UnitTests.Sagas;

public class SubscriptionSagaTests
{
    private readonly Mock<IPlanRepository> _planRepo = new();
    private readonly Mock<ISubscriptionRepository> _subRepo = new();
    private readonly Mock<IPaymentRepository> _payRepo = new();
    private readonly Mock<IPaymentGateway> _payGateway = new();
    private readonly Mock<IInvoiceRepository> _invoiceRepo = new();
    private readonly Mock<IOutboxRepository> _outboxRepo = new();
    private readonly Mock<ISagaRepository> _sagaRepo = new();
    private readonly Mock<ISubscriptionUnitOfWork> _unitOfWork = new();
    private readonly CompensatingActions _compensatingActions;
    private readonly SubscriptionSaga _saga;

    private readonly Guid _userId = Guid.NewGuid();
    private readonly Guid _planId = Guid.NewGuid();
    private const string CardNumber = "4242424242424242";

    private readonly Plan _standardPlan;

    public SubscriptionSagaTests()
    {
        _standardPlan = new Plan
        {
            Id = _planId,
            Name = "Standard",
            Tier = PlanTier.Standard,
            PriceMonthly = 79.99m,
            IsActive = true,
            MaxScreens = 3,
            MaxQuality = "1080p"
        };

        _unitOfWork.Setup(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync(1);

        _compensatingActions = new CompensatingActions(
            _subRepo.Object,
            _payGateway.Object,
            _sagaRepo.Object,
            _unitOfWork.Object,
            Mock.Of<ILogger<CompensatingActions>>());

        _saga = new SubscriptionSaga(
            _planRepo.Object,
            _subRepo.Object,
            _payRepo.Object,
            _payGateway.Object,
            _invoiceRepo.Object,
            _outboxRepo.Object,
            _sagaRepo.Object,
            _unitOfWork.Object,
            _compensatingActions,
            Mock.Of<ILogger<SubscriptionSaga>>());

        SetupHappyPath();
    }

    private void SetupHappyPath()
    {
        _planRepo.Setup(r => r.GetByIdAsync(_planId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_standardPlan);

        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync((Subscription?)null);

        _payGateway.Setup(g => g.ChargeAsync(CardNumber, _standardPlan.PriceMonthly, "TRY", It.IsAny<CancellationToken>()))
            .ReturnsAsync(new PaymentResult(true, "txn_123", null));

        _invoiceRepo.Setup(r => r.GenerateInvoiceNumberAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync("INV-2026-0001");
    }

    [Fact]
    public async Task ExecuteAsync_HappyPath_ReturnsActiveSubscription()
    {
        var result = await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        result.Should().NotBeNull();
        result.Status.Should().Be(SubscriptionStatus.Active);
        result.PlanId.Should().Be(_planId);
        result.PeriodEnd.Should().BeAfter(result.PeriodStart);
    }

    [Fact]
    public async Task ExecuteAsync_HappyPath_CreatesPaymentRecord()
    {
        await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        _payRepo.Verify(r => r.AddAsync(
            It.Is<Payment>(p =>
                p.Amount == _standardPlan.PriceMonthly &&
                p.Currency == "TRY" &&
                p.Status == PaymentStatus.Succeeded),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_HappyPath_CreatesInvoice()
    {
        await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        _invoiceRepo.Verify(r => r.AddAsync(
            It.Is<Invoice>(i => i.InvoiceNumber == "INV-2026-0001"),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_HappyPath_PublishesOutboxEvents()
    {
        await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        _outboxRepo.Verify(r => r.AddAsync(
            It.Is<OutboxMessage>(m => m.EventType == "subscription.created"),
            It.IsAny<CancellationToken>()), Times.Once);

        _outboxRepo.Verify(r => r.AddAsync(
            It.Is<OutboxMessage>(m => m.EventType == "subscription.payment.processed"),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_HappyPath_SagaCompletedStatus()
    {
        await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        _sagaRepo.Verify(r => r.UpdateAsync(
            It.Is<SagaState>(s => s.Status == SagaStatus.Completed),
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Exactly(2));
    }

    [Fact]
    public async Task ExecuteAsync_PlanNotFound_ThrowsPlanNotFoundException()
    {
        _planRepo.Setup(r => r.GetByIdAsync(_planId, It.IsAny<CancellationToken>()))
            .ReturnsAsync((Plan?)null);

        var act = () => _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<PlanNotFoundException>();
    }

    [Fact]
    public async Task ExecuteAsync_ActiveSubscriptionExists_ThrowsActiveSubscriptionExistsException()
    {
        var existing = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = _userId,
            Status = SubscriptionStatus.Active
        };
        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(existing);

        var act = () => _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<ActiveSubscriptionExistsException>();
    }

    [Fact]
    public async Task ExecuteAsync_PaymentFails_ThrowsPaymentFailedException()
    {
        _payGateway.Setup(g => g.ChargeAsync(CardNumber, _standardPlan.PriceMonthly, "TRY", It.IsAny<CancellationToken>()))
            .ReturnsAsync(new PaymentResult(false, null, "Insufficient funds"));

        var act = () => _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<PaymentFailedException>();
    }

    [Fact]
    public async Task ExecuteAsync_PaymentFails_SetsSubscriptionToFailed()
    {
        _payGateway.Setup(g => g.ChargeAsync(CardNumber, _standardPlan.PriceMonthly, "TRY", It.IsAny<CancellationToken>()))
            .ReturnsAsync(new PaymentResult(false, null, "Insufficient funds"));

        // Compensation calls GetByIdAsync to find and mark the subscription as Failed.
        // Capture the subscription that AddAsync creates so GetByIdAsync can return it.
        Subscription? addedSub = null;
        _subRepo.Setup(r => r.AddAsync(It.IsAny<Subscription>(), It.IsAny<CancellationToken>()))
            .Callback<Subscription, CancellationToken>((s, _) => addedSub = s)
            .Returns(Task.CompletedTask);
        _subRepo.Setup(r => r.GetByIdAsync(It.IsAny<Guid>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(() => addedSub);

        try
        {
            await _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);
        }
        catch (PaymentFailedException) { }

        // Subscription should be created first then set to Failed by compensation
        addedSub.Should().NotBeNull();
        addedSub!.Status.Should().Be(SubscriptionStatus.Failed);
    }

    [Fact]
    public async Task ExecuteAsync_InvoiceError_TriggersCompensation()
    {
        _invoiceRepo.Setup(r => r.GenerateInvoiceNumberAsync(It.IsAny<CancellationToken>()))
            .ThrowsAsync(new InvalidOperationException("DB error"));

        var act = () => _saga.ExecuteAsync(_userId, _planId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<PaymentFailedException>();

        // Compensation should have set saga to Failed
        _sagaRepo.Verify(r => r.UpdateAsync(
            It.Is<SagaState>(s => s.Status == SagaStatus.Failed),
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Exactly(3));
    }
}
