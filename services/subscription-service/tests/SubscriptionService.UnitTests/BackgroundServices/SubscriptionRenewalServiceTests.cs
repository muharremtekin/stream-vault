using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Moq;
using SubscriptionService.Api.BackgroundServices;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using Xunit;

namespace SubscriptionService.UnitTests.BackgroundServices;

public class SubscriptionRenewalServiceTests
{
    private readonly Mock<IServiceProvider> _serviceProviderMock = new();
    private readonly Mock<IServiceScope> _serviceScopeMock = new();
    private readonly Mock<IServiceScopeFactory> _scopeFactoryMock = new();
    private readonly Mock<ISubscriptionRepository> _subscriptionRepositoryMock = new();
    private readonly Mock<IOutboxRepository> _outboxRepositoryMock = new();
    private readonly Mock<ISubscriptionUnitOfWork> _unitOfWorkMock = new();
    private readonly Mock<ILogger<SubscriptionRenewalService>> _loggerMock = new();

    private readonly IOptions<SubscriptionRenewalOptions> _options = Options.Create(new SubscriptionRenewalOptions
    {
        IntervalSeconds = 1,
        BatchSize = 2
    });

    public SubscriptionRenewalServiceTests()
    {
        _scopeFactoryMock.Setup(factory => factory.CreateScope()).Returns(_serviceScopeMock.Object);

        _serviceProviderMock
            .Setup(provider => provider.GetService(typeof(IServiceScopeFactory)))
            .Returns(_scopeFactoryMock.Object);
    }

    private SubscriptionRenewalService CreateService() =>
        new(_serviceProviderMock.Object, _loggerMock.Object, _options);

    private void SetupScopedServices()
    {
        var scopedProviderMock = new Mock<IServiceProvider>();
        scopedProviderMock
            .Setup(provider => provider.GetService(typeof(ISubscriptionRepository)))
            .Returns(_subscriptionRepositoryMock.Object);
        scopedProviderMock
            .Setup(provider => provider.GetService(typeof(IOutboxRepository)))
            .Returns(_outboxRepositoryMock.Object);
        scopedProviderMock
            .Setup(provider => provider.GetService(typeof(ISubscriptionUnitOfWork)))
            .Returns(_unitOfWorkMock.Object);

        _serviceScopeMock.Setup(scope => scope.ServiceProvider).Returns(scopedProviderMock.Object);
    }

    [Fact]
    public async Task ProcessExpiredSubscriptions_ProcessesBatchWithSingleCommitForRenewals()
    {
        var dueSubscriptions = new List<ExpiredSubscriptionBatchItem>
        {
            new()
            {
                Id = Guid.NewGuid(),
                UserId = Guid.NewGuid(),
                PlanId = Guid.NewGuid(),
                PeriodEnd = DateTime.UtcNow.AddMinutes(-5),
                AutoRenew = true
            },
            new()
            {
                Id = Guid.NewGuid(),
                UserId = Guid.NewGuid(),
                PlanId = Guid.NewGuid(),
                PeriodEnd = DateTime.UtcNow.AddMinutes(-10),
                AutoRenew = false
            }
        };

        SetupScopedServices();
        _subscriptionRepositoryMock
            .SetupSequence(repository => repository.GetExpiredSubscriptionsAsync(2, It.IsAny<CancellationToken>()))
            .ReturnsAsync(dueSubscriptions)
            .ReturnsAsync([])
            .ReturnsAsync([]);
        _subscriptionRepositoryMock
            .Setup(repository => repository.ExpireSubscriptionsAsync(
                It.IsAny<IReadOnlyCollection<Guid>>(),
                It.IsAny<DateTime>(),
                It.IsAny<CancellationToken>()))
            .ReturnsAsync(1);
        _subscriptionRepositoryMock
            .Setup(repository => repository.UpdateRenewalBatchAsync(
                It.IsAny<IReadOnlyCollection<SubscriptionRenewalUpdate>>(),
                It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _outboxRepositoryMock
            .Setup(repository => repository.AddAsync(It.IsAny<OutboxMessage>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _unitOfWorkMock
            .Setup(unitOfWork => unitOfWork.SaveChangesAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync(2);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _subscriptionRepositoryMock.Verify(repository => repository.GetExpiredSubscriptionsAsync(2, It.IsAny<CancellationToken>()), Times.AtLeast(2));
        _subscriptionRepositoryMock.Verify(repository => repository.ExpireSubscriptionsAsync(
            It.Is<IReadOnlyCollection<Guid>>(ids => ids.Count == 1 && ids.Contains(dueSubscriptions[1].Id)),
            It.IsAny<DateTime>(),
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _subscriptionRepositoryMock.Verify(repository => repository.UpdateRenewalBatchAsync(
            It.Is<IReadOnlyCollection<SubscriptionRenewalUpdate>>(renewals =>
                renewals.Count == 1
                && renewals.Single().SubscriptionId == dueSubscriptions[0].Id
                && renewals.Single().NewPeriodStart == dueSubscriptions[0].PeriodEnd
                && renewals.Single().NewPeriodEnd == dueSubscriptions[0].PeriodEnd.AddDays(30)),
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _outboxRepositoryMock.Verify(repository => repository.AddAsync(
            It.Is<OutboxMessage>(message => message.EventType == "subscription.renewed"),
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _unitOfWorkMock.Verify(unitOfWork => unitOfWork.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.AtLeastOnce);
    }

    [Fact]
    public async Task ProcessExpiredSubscriptions_FullBatchFetchesNextBatchWithinSameCycle()
    {
        var firstBatch = new List<ExpiredSubscriptionBatchItem>
        {
            new()
            {
                Id = Guid.NewGuid(),
                UserId = Guid.NewGuid(),
                PlanId = Guid.NewGuid(),
                PeriodEnd = DateTime.UtcNow.AddMinutes(-10),
                AutoRenew = true
            },
            new()
            {
                Id = Guid.NewGuid(),
                UserId = Guid.NewGuid(),
                PlanId = Guid.NewGuid(),
                PeriodEnd = DateTime.UtcNow.AddMinutes(-5),
                AutoRenew = true
            }
        };
        var secondBatch = new List<ExpiredSubscriptionBatchItem>
        {
            new()
            {
                Id = Guid.NewGuid(),
                UserId = Guid.NewGuid(),
                PlanId = Guid.NewGuid(),
                PeriodEnd = DateTime.UtcNow.AddMinutes(-1),
                AutoRenew = true
            }
        };

        SetupScopedServices();
        _subscriptionRepositoryMock
            .SetupSequence(repository => repository.GetExpiredSubscriptionsAsync(2, It.IsAny<CancellationToken>()))
            .ReturnsAsync(firstBatch)
            .ReturnsAsync(secondBatch)
            .ReturnsAsync([])
            .ReturnsAsync([]);
        _subscriptionRepositoryMock
            .Setup(repository => repository.UpdateRenewalBatchAsync(
                It.IsAny<IReadOnlyCollection<SubscriptionRenewalUpdate>>(),
                It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _outboxRepositoryMock
            .Setup(repository => repository.AddAsync(It.IsAny<OutboxMessage>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _unitOfWorkMock
            .Setup(unitOfWork => unitOfWork.SaveChangesAsync(It.IsAny<CancellationToken>()))
            .ReturnsAsync(2);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _subscriptionRepositoryMock.Verify(repository => repository.GetExpiredSubscriptionsAsync(2, It.IsAny<CancellationToken>()), Times.AtLeast(2));
        _subscriptionRepositoryMock.Verify(repository => repository.UpdateRenewalBatchAsync(
            It.IsAny<IReadOnlyCollection<SubscriptionRenewalUpdate>>(),
            It.IsAny<CancellationToken>()), Times.AtLeast(2));
        _unitOfWorkMock.Verify(unitOfWork => unitOfWork.SaveChangesAsync(It.IsAny<CancellationToken>()), Times.AtLeast(2));
    }
}
