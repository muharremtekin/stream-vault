using FluentAssertions;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Application.Sagas;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;

namespace SubscriptionService.UnitTests.Sagas;

public class CompensatingActionsTests
{
    private readonly Mock<ISubscriptionRepository> _subRepo = new();
    private readonly Mock<IPaymentGateway> _payGateway = new();
    private readonly Mock<ISagaRepository> _sagaRepo = new();
    private readonly Mock<ISubscriptionUnitOfWork> _unitOfWork = new();
    private readonly CompensatingActions _compensatingActions;

    public CompensatingActionsTests()
    {
        _unitOfWork.Setup(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync(1);

        _compensatingActions = new CompensatingActions(
            _subRepo.Object,
            _payGateway.Object,
            _sagaRepo.Object,
            _unitOfWork.Object,
            Mock.Of<ILogger<CompensatingActions>>());
    }

    private SagaState CreateSaga() => new()
    {
        Id = Guid.NewGuid(),
        SagaType = "CreateSubscription",
        CurrentStep = SagaStep.ProcessPayment,
        Status = SagaStatus.InProgress
    };

    [Fact]
    public async Task CompensateCreateSubscription_WithTransactionId_CallsRefund()
    {
        var saga = CreateSaga();
        var data = new CreateSubscriptionSagaData { TransactionId = "txn_123" };

        await _compensatingActions.CompensateCreateSubscriptionAsync(saga, data, CancellationToken.None);

        _payGateway.Verify(g => g.RefundAsync("txn_123", 0, It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task CompensateCreateSubscription_WithoutTransactionId_SkipsRefund()
    {
        var saga = CreateSaga();
        var data = new CreateSubscriptionSagaData { TransactionId = null };

        await _compensatingActions.CompensateCreateSubscriptionAsync(saga, data, CancellationToken.None);

        _payGateway.Verify(g => g.RefundAsync(
            It.IsAny<string>(), It.IsAny<decimal>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task CompensateCreateSubscription_WithSubscriptionId_SetsStatusFailed()
    {
        var saga = CreateSaga();
        var subId = Guid.NewGuid();
        var subscription = new Subscription
        {
            Id = subId,
            Status = SubscriptionStatus.PendingPayment
        };
        var data = new CreateSubscriptionSagaData { SubscriptionId = subId };

        _subRepo.Setup(r => r.GetByIdAsync(subId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(subscription);

        await _compensatingActions.CompensateCreateSubscriptionAsync(saga, data, CancellationToken.None);

        subscription.Status.Should().Be(SubscriptionStatus.Failed);
        _subRepo.Verify(r => r.UpdateAsync(
            It.Is<Subscription>(s => s.Status == SubscriptionStatus.Failed),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task CompensateCreateSubscription_SetsSagaToFailed()
    {
        var saga = CreateSaga();
        var data = new CreateSubscriptionSagaData();

        await _compensatingActions.CompensateCreateSubscriptionAsync(saga, data, CancellationToken.None);

        saga.Status.Should().Be(SagaStatus.Failed);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task CompensateChangePlan_WithTransactionId_RefundsWithAmount()
    {
        var saga = CreateSaga();
        saga.SagaType = "ChangePlan";
        var data = new ChangePlanSagaData
        {
            TransactionId = "txn_upgrade",
            PriceDifference = 25.50m
        };

        await _compensatingActions.CompensateChangePlanAsync(saga, data, CancellationToken.None);

        _payGateway.Verify(g => g.RefundAsync("txn_upgrade", 25.50m, It.IsAny<CancellationToken>()), Times.Once);
        saga.Status.Should().Be(SagaStatus.Failed);
        _unitOfWork.Verify(u => u.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.Exactly(2));
    }
}
