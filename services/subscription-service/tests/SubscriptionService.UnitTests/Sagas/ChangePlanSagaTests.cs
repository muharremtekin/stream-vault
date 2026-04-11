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

public class ChangePlanSagaTests
{
    private readonly Mock<IPlanRepository> _planRepo = new();
    private readonly Mock<ISubscriptionRepository> _subRepo = new();
    private readonly Mock<IPaymentRepository> _payRepo = new();
    private readonly Mock<IPaymentGateway> _payGateway = new();
    private readonly Mock<IOutboxRepository> _outboxRepo = new();
    private readonly Mock<ISagaRepository> _sagaRepo = new();
    private readonly Mock<ISubscriptionUnitOfWork> _unitOfWork = new();
    private readonly CompensatingActions _compensatingActions;
    private readonly ChangePlanSaga _saga;

    private readonly Guid _subscriptionId = Guid.NewGuid();
    private readonly Guid _standardPlanId = Guid.NewGuid();
    private readonly Guid _premiumPlanId = Guid.NewGuid();
    private readonly Guid _basicPlanId = Guid.NewGuid();
    private readonly Guid _userId = Guid.NewGuid();
    private const string CardNumber = "4242424242424242";

    private readonly Plan _standardPlan;
    private readonly Plan _premiumPlan;
    private readonly Plan _basicPlan;
    private readonly Subscription _activeSubscription;

    public ChangePlanSagaTests()
    {
        _standardPlan = new Plan
        {
            Id = _standardPlanId,
            Name = "Standard",
            Tier = PlanTier.Standard,
            PriceMonthly = 79.99m,
            IsActive = true
        };

        _premiumPlan = new Plan
        {
            Id = _premiumPlanId,
            Name = "Premium",
            Tier = PlanTier.Premium,
            PriceMonthly = 119.99m,
            IsActive = true
        };

        _basicPlan = new Plan
        {
            Id = _basicPlanId,
            Name = "Basic",
            Tier = PlanTier.Basic,
            PriceMonthly = 49.99m,
            IsActive = true
        };

        _activeSubscription = new Subscription
        {
            Id = _subscriptionId,
            UserId = _userId,
            PlanId = _standardPlanId,
            Plan = _standardPlan,
            Status = SubscriptionStatus.Active,
            PeriodStart = DateTime.UtcNow.AddDays(-5),
            PeriodEnd = DateTime.UtcNow.AddDays(25),
            AutoRenew = true
        };

        _unitOfWork.Setup(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync(1);

        _compensatingActions = new CompensatingActions(
            _subRepo.Object,
            _payGateway.Object,
            _sagaRepo.Object,
            _unitOfWork.Object,
            Mock.Of<ILogger<CompensatingActions>>());

        _saga = new ChangePlanSaga(
            _planRepo.Object,
            _subRepo.Object,
            _payRepo.Object,
            _payGateway.Object,
            _outboxRepo.Object,
            _sagaRepo.Object,
            _unitOfWork.Object,
            _compensatingActions,
            Mock.Of<ILogger<ChangePlanSaga>>());
    }

    private void SetupUpgrade()
    {
        _subRepo.Setup(r => r.GetByIdAsync(_subscriptionId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_activeSubscription);
        _planRepo.Setup(r => r.GetByIdAsync(_premiumPlanId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_premiumPlan);
        _payGateway.Setup(g => g.ChargeAsync(CardNumber, It.IsAny<decimal>(), "TRY", It.IsAny<CancellationToken>()))
            .ReturnsAsync(new PaymentResult(true, "txn_upgrade", null));
    }

    private void SetupDowngrade()
    {
        _subRepo.Setup(r => r.GetByIdAsync(_subscriptionId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_activeSubscription);
        _planRepo.Setup(r => r.GetByIdAsync(_basicPlanId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_basicPlan);
    }

    [Fact]
    public async Task ExecuteAsync_Upgrade_ChargesPositivePriceDifference()
    {
        SetupUpgrade();

        await _saga.ExecuteAsync(_subscriptionId, _premiumPlanId, CardNumber, CancellationToken.None);

        _payGateway.Verify(g => g.ChargeAsync(
            CardNumber,
            It.Is<decimal>(amount => amount > 0),
            "TRY",
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_Upgrade_UpdatesSubscriptionPlan()
    {
        SetupUpgrade();

        var result = await _saga.ExecuteAsync(_subscriptionId, _premiumPlanId, CardNumber, CancellationToken.None);

        result.PlanId.Should().Be(_premiumPlanId);
    }

    [Fact]
    public async Task ExecuteAsync_Upgrade_PublishesPlanChangedEvent()
    {
        SetupUpgrade();

        await _saga.ExecuteAsync(_subscriptionId, _premiumPlanId, CardNumber, CancellationToken.None);

        _outboxRepo.Verify(r => r.AddAsync(
            It.Is<OutboxMessage>(m => m.EventType == "plan.changed"),
            It.IsAny<CancellationToken>()), Times.Once);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Exactly(2));
    }

    [Fact]
    public async Task ExecuteAsync_Downgrade_CreatesNegativePayment()
    {
        SetupDowngrade();

        await _saga.ExecuteAsync(_subscriptionId, _basicPlanId, CardNumber, CancellationToken.None);

        _payRepo.Verify(r => r.AddAsync(
            It.Is<Payment>(p => p.Amount < 0 && p.Status == PaymentStatus.Succeeded),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_Downgrade_DoesNotCallPaymentGateway()
    {
        SetupDowngrade();

        await _saga.ExecuteAsync(_subscriptionId, _basicPlanId, CardNumber, CancellationToken.None);

        _payGateway.Verify(g => g.ChargeAsync(
            It.IsAny<string>(), It.IsAny<decimal>(), It.IsAny<string>(), It.IsAny<CancellationToken>()),
            Times.Never);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ExecuteAsync_SubscriptionNotFound_ThrowsSubscriptionNotFoundException()
    {
        _subRepo.Setup(r => r.GetByIdAsync(_subscriptionId, It.IsAny<CancellationToken>()))
            .ReturnsAsync((Subscription?)null);

        var act = () => _saga.ExecuteAsync(_subscriptionId, _premiumPlanId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<SubscriptionNotFoundException>();
    }

    [Fact]
    public async Task ExecuteAsync_SamePlan_ThrowsInvalidOperationException()
    {
        _subRepo.Setup(r => r.GetByIdAsync(_subscriptionId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_activeSubscription);
        // Same plan as current
        _planRepo.Setup(r => r.GetByIdAsync(_standardPlanId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_standardPlan);

        var act = () => _saga.ExecuteAsync(_subscriptionId, _standardPlanId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<InvalidOperationException>()
            .WithMessage("*already on this plan*");
    }

    [Fact]
    public async Task ExecuteAsync_UpgradePaymentFails_ThrowsPaymentFailedException()
    {
        _subRepo.Setup(r => r.GetByIdAsync(_subscriptionId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_activeSubscription);
        _planRepo.Setup(r => r.GetByIdAsync(_premiumPlanId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(_premiumPlan);
        _payGateway.Setup(g => g.ChargeAsync(CardNumber, It.IsAny<decimal>(), "TRY", It.IsAny<CancellationToken>()))
            .ReturnsAsync(new PaymentResult(false, null, "Card declined"));

        var act = () => _saga.ExecuteAsync(_subscriptionId, _premiumPlanId, CardNumber, CancellationToken.None);

        await act.Should().ThrowAsync<PaymentFailedException>();
    }
}
