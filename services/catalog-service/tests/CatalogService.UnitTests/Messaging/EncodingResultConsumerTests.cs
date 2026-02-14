using CatalogService.Application.Commands.UpdateVideoStatus;
using CatalogService.Domain.Enums;
using CatalogService.Infrastructure.Messaging;
using MediatR;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;

namespace CatalogService.UnitTests.Messaging;

public class EncodingResultConsumerTests
{
    private readonly Mock<IMediator> _mediatorMock;
    private readonly Mock<ILogger<EncodingResultConsumer>> _loggerMock;
    private readonly EncodingResultConsumer _consumer;

    public EncodingResultConsumerTests()
    {
        _mediatorMock = new Mock<IMediator>();
        _loggerMock = new Mock<ILogger<EncodingResultConsumer>>();

        // Setup IServiceScopeFactory to return a scope with our mock mediator
        var serviceProviderMock = new Mock<IServiceProvider>();
        serviceProviderMock
            .Setup(sp => sp.GetService(typeof(IMediator)))
            .Returns(_mediatorMock.Object);

        var scopeMock = new Mock<IServiceScope>();
        scopeMock.Setup(s => s.ServiceProvider).Returns(serviceProviderMock.Object);

        var scopeFactoryMock = new Mock<IServiceScopeFactory>();
        scopeFactoryMock.Setup(f => f.CreateScope()).Returns(scopeMock.Object);

        // Setup configuration with a dummy connection string
        var config = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["RabbitMQ:ConnectionString"] = "amqp://guest:guest@localhost:5672/",
                ["RabbitMQ:QueueName"] = "encoding.results.catalog"
            })
            .Build();

        _consumer = new EncodingResultConsumer(
            scopeFactoryMock.Object,
            _loggerMock.Object,
            config);
    }

    [Fact]
    public async Task HandleMessage_CompletedStatus_SendsReadyCommand()
    {
        var message = CreateCompletedMessage();

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(true);

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        _mediatorMock.Verify(m => m.Send(
            It.Is<UpdateVideoStatusCommand>(cmd =>
                cmd.ContentId == "movie-123" &&
                cmd.VideoStatus == VideoStatus.Ready &&
                cmd.StreamingInfo != null),
            It.IsAny<CancellationToken>()),
            Times.Once);
    }

    [Fact]
    public async Task HandleMessage_CompletedStatus_MapsStreamingInfoCorrectly()
    {
        var message = CreateCompletedMessage();
        UpdateVideoStatusCommand? capturedCommand = null;

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .Callback<IRequest<bool>, CancellationToken>((cmd, _) =>
                capturedCommand = cmd as UpdateVideoStatusCommand)
            .ReturnsAsync(true);

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        Assert.NotNull(capturedCommand);
        Assert.NotNull(capturedCommand!.StreamingInfo);

        var info = capturedCommand.StreamingInfo!;
        Assert.Equal(120, info.DurationSeconds);
        Assert.Equal("movie-123/manifest.m3u8", info.ManifestPath);
        Assert.Equal("movie-123/thumb_300x170.jpg", info.ThumbnailPath);
        Assert.Equal("movie-123/poster.jpg", info.PosterPath);
    }

    [Fact]
    public async Task HandleMessage_CompletedStatus_MapsQualityInfoCorrectly()
    {
        var message = CreateCompletedMessage();
        UpdateVideoStatusCommand? capturedCommand = null;

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .Callback<IRequest<bool>, CancellationToken>((cmd, _) =>
                capturedCommand = cmd as UpdateVideoStatusCommand)
            .ReturnsAsync(true);

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        Assert.NotNull(capturedCommand?.StreamingInfo);
        var qualities = capturedCommand!.StreamingInfo!.AvailableQualities;
        Assert.Equal(2, qualities.Count);

        Assert.Equal("720p", qualities[0].Label);
        Assert.Equal(1280, qualities[0].Width);
        Assert.Equal(720, qualities[0].Height);
        Assert.Equal(2800, qualities[0].BitrateKbps);
        Assert.Equal(12, qualities[0].SegmentCount);

        Assert.Equal("1080p", qualities[1].Label);
        Assert.Equal(1920, qualities[1].Width);
        Assert.Equal(1080, qualities[1].Height);
        Assert.Equal(5000, qualities[1].BitrateKbps);
        Assert.Equal(15, qualities[1].SegmentCount);
    }

    [Fact]
    public async Task HandleMessage_FailedStatus_SendsErrorCommand()
    {
        var message = CreateFailedMessage();

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(true);

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        _mediatorMock.Verify(m => m.Send(
            It.Is<UpdateVideoStatusCommand>(cmd =>
                cmd.ContentId == "movie-123" &&
                cmd.VideoStatus == VideoStatus.Error &&
                cmd.StreamingInfo == null),
            It.IsAny<CancellationToken>()),
            Times.Once);
    }

    [Fact]
    public async Task HandleMessage_UnknownStatus_DoesNotSendCommand()
    {
        var message = new EncodingResultMessage
        {
            EventId = "evt-001",
            EventType = "encoding.job.unknown",
            ContentId = "movie-123",
            Status = "unknown_status",
            Outputs = new List<EncodingOutputMessage>(),
            CompletedAt = "2024-06-15T10:30:00Z"
        };

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        _mediatorMock.Verify(m => m.Send(
            It.IsAny<UpdateVideoStatusCommand>(),
            It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task HandleMessage_CompletedStatus_ParsesCompletedAtTimestamp()
    {
        var message = CreateCompletedMessage();
        message.CompletedAt = "2024-06-15T10:30:00Z";
        UpdateVideoStatusCommand? capturedCommand = null;

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .Callback<IRequest<bool>, CancellationToken>((cmd, _) =>
                capturedCommand = cmd as UpdateVideoStatusCommand)
            .ReturnsAsync(true);

        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        Assert.NotNull(capturedCommand?.StreamingInfo);
        Assert.Equal(
            new DateTime(2024, 6, 15, 10, 30, 0, DateTimeKind.Utc),
            capturedCommand!.StreamingInfo!.EncodedAt);
    }

    [Fact]
    public async Task HandleMessage_MediatorReturnsFalse_DoesNotThrow()
    {
        var message = CreateCompletedMessage();

        _mediatorMock
            .Setup(m => m.Send(It.IsAny<UpdateVideoStatusCommand>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(false);

        // Should not throw even when mediator returns false (movie not found)
        await _consumer.HandleMessageAsync(message, CancellationToken.None);

        _mediatorMock.Verify(m => m.Send(
            It.IsAny<UpdateVideoStatusCommand>(),
            It.IsAny<CancellationToken>()),
            Times.Once);
    }

    private static EncodingResultMessage CreateCompletedMessage() => new()
    {
        EventId = "evt-001",
        EventType = "encoding.job.completed",
        Timestamp = "2024-06-15T10:30:00Z",
        Source = "encoding-service",
        CorrelationId = "job-001",
        JobId = "job-001",
        ContentId = "movie-123",
        Status = "completed",
        DurationSeconds = 120,
        CompletedAt = "2024-06-15T10:30:00Z",
        Outputs = new List<EncodingOutputMessage>
        {
            new()
            {
                Quality = "720p",
                Width = 1280,
                Height = 720,
                BitrateKbps = 2800,
                SegmentCount = 12,
                PlaylistPath = "movie-123/720p/playlist.m3u8"
            },
            new()
            {
                Quality = "1080p",
                Width = 1920,
                Height = 1080,
                BitrateKbps = 5000,
                SegmentCount = 15,
                PlaylistPath = "movie-123/1080p/playlist.m3u8"
            }
        }
    };

    private static EncodingResultMessage CreateFailedMessage() => new()
    {
        EventId = "evt-002",
        EventType = "encoding.job.failed",
        Timestamp = "2024-06-15T10:30:00Z",
        Source = "encoding-service",
        CorrelationId = "job-001",
        JobId = "job-001",
        ContentId = "movie-123",
        Status = "failed",
        DurationSeconds = 0,
        ErrorMessage = "unsupported codec",
        CompletedAt = "2024-06-15T10:30:00Z",
        Outputs = new List<EncodingOutputMessage>()
    };
}
