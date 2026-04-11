using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Moq;
using SubscriptionService.Api.BackgroundServices;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Domain.Entities;
using FluentAssertions;
using Xunit;

namespace SubscriptionService.UnitTests.BackgroundServices;

public class OutboxProcessorServiceTests
{
    private readonly Mock<IServiceProvider> _serviceProviderMock = new();
    private readonly Mock<IServiceScope> _serviceScopeMock = new();
    private readonly Mock<IServiceScopeFactory> _scopeFactoryMock = new();
    private readonly Mock<IOutboxRepository> _outboxRepositoryMock = new();
    private readonly Mock<IRabbitMqPublisher> _publisherMock = new();
    private readonly Mock<ILogger<OutboxProcessorService>> _loggerMock = new();

    private readonly IOptions<OutboxProcessorOptions> _options = Options.Create(new OutboxProcessorOptions
    {
        IntervalSeconds = 1,
        MaxRetryCount = 3,
        BatchSize = 10,
        ShutdownTimeoutSeconds = 5
    });

    public OutboxProcessorServiceTests()
    {
        _serviceScopeMock.Setup(s => s.ServiceProvider).Returns(_serviceProviderMock.Object);
        _scopeFactoryMock.Setup(f => f.CreateScope()).Returns(_serviceScopeMock.Object);

        // IServiceProvider.CreateScope() -> IServiceScopeFactory üzerinden çalışır
        _serviceProviderMock
            .Setup(p => p.GetService(typeof(IServiceScopeFactory)))
            .Returns(_scopeFactoryMock.Object);

        _serviceProviderMock
            .Setup(p => p.GetService(typeof(IOutboxRepository)))
            .Returns(_outboxRepositoryMock.Object);

        _publisherMock.Setup(p => p.IsConnectedAsync()).ReturnsAsync(true);
    }

    private OutboxProcessorService CreateService() =>
        new(_serviceProviderMock.Object, _publisherMock.Object, _loggerMock.Object, _options);

    // ── Helper: scope içindeki IOutboxRepository'yi bağlar ──────────────────
    private void SetupScopedRepository()
    {
        var scopedProviderMock = new Mock<IServiceProvider>();
        scopedProviderMock
            .Setup(p => p.GetService(typeof(IOutboxRepository)))
            .Returns(_outboxRepositoryMock.Object);
        _serviceScopeMock.Setup(s => s.ServiceProvider).Returns(scopedProviderMock.Object);
    }

    // ────────────────────────────────────────────────────────────────────────
    // 1. Başarılı publish → MarkAsProcessedAsync çağrılır
    // ────────────────────────────────────────────────────────────────────────
    [Fact]
    public async Task ProcessMessages_PublishesSuccessfully_MarksAsProcessed()
    {
        var message = new OutboxMessage
        {
            Id = Guid.NewGuid(),
            EventType = "SubscriptionCreated",
            Payload = "{}",
            RetryCount = 0
        };

        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([message]);
        _publisherMock.Setup(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _outboxRepositoryMock.Setup(r => r.MarkAsProcessedAsync(message.Id, It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        // Tek bir döngü çalıştır, sonra iptal et
        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _outboxRepositoryMock.Verify(r => r.MarkAsProcessedAsync(message.Id, It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _outboxRepositoryMock.Verify(r => r.IncrementRetryAsync(It.IsAny<Guid>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    // ────────────────────────────────────────────────────────────────────────
    // 2. Publish başarısız → IncrementRetryAsync çağrılır
    // ────────────────────────────────────────────────────────────────────────
    [Fact]
    public async Task ProcessMessages_PublishFails_IncrementsRetry()
    {
        var message = new OutboxMessage
        {
            Id = Guid.NewGuid(),
            EventType = "PlanChanged",
            Payload = "{}",
            RetryCount = 0
        };

        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([message]);
        _publisherMock.Setup(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .ThrowsAsync(new InvalidOperationException("RabbitMQ down"));
        _outboxRepositoryMock.Setup(r => r.IncrementRetryAsync(message.Id, It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _outboxRepositoryMock.Verify(r => r.IncrementRetryAsync(message.Id, It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _outboxRepositoryMock.Verify(r => r.MarkAsProcessedAsync(It.IsAny<Guid>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    [Fact]
    public async Task ProcessMessages_DueRetryMessageReturnedByRepository_IsPublishedWithoutInMemoryBackoffSkip()
    {
        var message = new OutboxMessage
        {
            Id = Guid.NewGuid(),
            EventType = "RetryReady",
            Payload = "{}",
            RetryCount = 2,
            LastAttemptedAt = DateTime.UtcNow.AddMinutes(10),
            NextAttemptAt = DateTime.UtcNow.AddMinutes(-1)
        };

        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([message]);
        _publisherMock.Setup(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _outboxRepositoryMock.Setup(r => r.MarkAsProcessedAsync(message.Id, It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _publisherMock.Verify(p => p.PublishAsync(
            It.IsAny<string>(),
            message.EventType,
            message.Payload,
            It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _outboxRepositoryMock.Verify(r => r.MarkAsProcessedAsync(message.Id, It.IsAny<CancellationToken>()), Times.AtLeastOnce);
    }

    [Fact]
    public async Task ProcessMessages_FullBatch_ContinuesFetchingDueMessagesWithinSameCycle()
    {
        var firstBatch = Enumerable.Range(0, 10)
            .Select(i => new OutboxMessage
            {
                Id = Guid.NewGuid(),
                EventType = $"Event{i}",
                Payload = "{}",
                RetryCount = 0
            })
            .ToList();
        var secondBatch = new List<OutboxMessage>
        {
            new()
            {
                Id = Guid.NewGuid(),
                EventType = "Event10",
                Payload = "{}",
                RetryCount = 0
            }
        };

        SetupScopedRepository();
        _outboxRepositoryMock.SetupSequence(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync(firstBatch)
            .ReturnsAsync(secondBatch)
            .ReturnsAsync([])
            .ReturnsAsync([]);
        _publisherMock.Setup(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);
        _outboxRepositoryMock.Setup(r => r.MarkAsProcessedAsync(It.IsAny<Guid>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _outboxRepositoryMock.Verify(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()), Times.AtLeast(2));
        _publisherMock.Verify(p => p.PublishAsync(
            It.IsAny<string>(),
            It.IsAny<string>(),
            It.IsAny<string>(),
            It.IsAny<CancellationToken>()), Times.AtLeast(11));
    }

    // ────────────────────────────────────────────────────────────────────────
    // 3. MaxRetry aşıldı → MarkAsDeadLetterAsync çağrılır, publish yapılmaz
    // ────────────────────────────────────────────────────────────────────────
    [Fact]
    public async Task ProcessMessages_MaxRetryExceeded_MarksAsDeadLetter()
    {
        var message = new OutboxMessage
        {
            Id = Guid.NewGuid(),
            EventType = "SubscriptionCancelled",
            Payload = "{}",
            RetryCount = 3 // MaxRetryCount = 3 ile eşit → dead-letter
        };

        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([message]);
        _outboxRepositoryMock.Setup(r => r.MarkAsDeadLetterAsync(message.Id, It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _outboxRepositoryMock.Verify(r => r.MarkAsDeadLetterAsync(message.Id, It.IsAny<CancellationToken>()), Times.AtLeastOnce);
        _publisherMock.Verify(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    // ────────────────────────────────────────────────────────────────────────
    // 4. Kuyruk boş → hiçbir şey yapılmaz
    // ────────────────────────────────────────────────────────────────────────
    [Fact]
    public async Task ProcessMessages_EmptyQueue_DoesNothing()
    {
        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([]);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _publisherMock.Verify(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
        _outboxRepositoryMock.Verify(r => r.MarkAsProcessedAsync(It.IsAny<Guid>(), It.IsAny<CancellationToken>()), Times.Never);
        _outboxRepositoryMock.Verify(r => r.IncrementRetryAsync(It.IsAny<Guid>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
        _outboxRepositoryMock.Verify(r => r.MarkAsDeadLetterAsync(It.IsAny<Guid>(), It.IsAny<CancellationToken>()), Times.Never);
    }

    // ────────────────────────────────────────────────────────────────────────
    // 5. RabbitMQ bağlantısı yoksa işlem yapılmaz (IsConnectedAsync = false)
    // ────────────────────────────────────────────────────────────────────────
    [Fact]
    public async Task ProcessMessages_RabbitMqDisconnected_SkipsProcessing()
    {
        _publisherMock.Setup(p => p.IsConnectedAsync()).ReturnsAsync(false);

        var message = new OutboxMessage { Id = Guid.NewGuid(), EventType = "Test", Payload = "{}", RetryCount = 0 };
        SetupScopedRepository();
        _outboxRepositoryMock.Setup(r => r.GetUnprocessedAsync(10, It.IsAny<CancellationToken>()))
            .ReturnsAsync([message]);

        using var cts = new CancellationTokenSource();
        var service = CreateService();

        var task = service.StartAsync(cts.Token);
        await Task.Delay(200);
        await cts.CancelAsync();
        await task;

        _publisherMock.Verify(p => p.PublishAsync(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<CancellationToken>()), Times.Never);
    }
}
