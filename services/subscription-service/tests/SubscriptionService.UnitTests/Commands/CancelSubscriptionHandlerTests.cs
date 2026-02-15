using FluentAssertions;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;
using SubscriptionService.Application.Commands.CancelSubscription;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Exceptions;

namespace SubscriptionService.UnitTests.Commands;

public class CancelSubscriptionHandlerTests
{
    private readonly Mock<ISubscriptionRepository> _subRepo = new();
    private readonly Mock<IOutboxRepository> _outboxRepo = new();
    private readonly CancelSubscriptionHandler _handler;

    private readonly Guid _userId = Guid.NewGuid();

    public CancelSubscriptionHandlerTests()
    {
        _handler = new CancelSubscriptionHandler(
            _subRepo.Object,
            _outboxRepo.Object,
            Mock.Of<ILogger<CancelSubscriptionHandler>>());
    }

    [Fact]
    public async Task Handle_ActiveSubscription_MarksCancelled()
    {
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = _userId,
            Status = SubscriptionStatus.Active,
            PeriodEnd = DateTime.UtcNow.AddDays(25),
            AutoRenew = true
        };
        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(subscription);

        var command = new CancelSubscriptionCommand(_userId);
        var result = await _handler.Handle(command, CancellationToken.None);

        result.Status.Should().Be("Cancelled");
        result.CancelledAt.Should().BeCloseTo(DateTime.UtcNow, TimeSpan.FromSeconds(5));
        subscription.AutoRenew.Should().BeFalse();
    }

    [Fact]
    public async Task Handle_ActiveSubscription_PublishesOutboxEvent()
    {
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = _userId,
            Status = SubscriptionStatus.Active,
            PeriodEnd = DateTime.UtcNow.AddDays(25)
        };
        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(subscription);

        var command = new CancelSubscriptionCommand(_userId);
        await _handler.Handle(command, CancellationToken.None);

        _outboxRepo.Verify(r => r.AddAsync(
            It.Is<OutboxMessage>(m => m.EventType == "subscription.cancelled"),
            It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task Handle_NoActiveSubscription_ThrowsSubscriptionNotFoundException()
    {
        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync((Subscription?)null);

        var command = new CancelSubscriptionCommand(_userId);
        var act = () => _handler.Handle(command, CancellationToken.None);

        await act.Should().ThrowAsync<SubscriptionNotFoundException>();
    }

    [Fact]
    public async Task Handle_NonActiveSubscription_ThrowsInvalidOperationException()
    {
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = _userId,
            Status = SubscriptionStatus.Cancelled
        };
        _subRepo.Setup(r => r.GetActiveByUserIdAsync(_userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(subscription);

        var command = new CancelSubscriptionCommand(_userId);
        var act = () => _handler.Handle(command, CancellationToken.None);

        await act.Should().ThrowAsync<InvalidOperationException>();
    }
}
