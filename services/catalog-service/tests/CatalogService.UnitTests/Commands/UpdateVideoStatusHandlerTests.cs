using CatalogService.Application.Commands.UpdateVideoStatus;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using Microsoft.Extensions.Logging;
using Moq;
using Xunit;

namespace CatalogService.UnitTests.Commands;

public class UpdateVideoStatusHandlerTests
{
    private readonly Mock<IMovieRepository> _movieRepositoryMock;
    private readonly Mock<ILogger<UpdateVideoStatusHandler>> _loggerMock;
    private readonly UpdateVideoStatusHandler _handler;

    public UpdateVideoStatusHandlerTests()
    {
        _movieRepositoryMock = new Mock<IMovieRepository>();
        _loggerMock = new Mock<ILogger<UpdateVideoStatusHandler>>();
        _handler = new UpdateVideoStatusHandler(_movieRepositoryMock.Object, _loggerMock.Object);
    }

    private static Movie CreateMovie(VideoStatus status = VideoStatus.NotUploaded) => new()
    {
        Id = "507f1f77bcf86cd799439011",
        Title = "Test Movie",
        Description = "Test",
        ReleaseYear = 2024,
        Director = "Director",
        ThumbnailUrl = "https://example.com/thumb.jpg",
        BannerUrl = "https://example.com/banner.jpg",
        Genres = new List<string> { "Action" },
        Cast = new List<CastMember>(),
        VideoStatus = status
    };

    private static StreamingInfo CreateStreamingInfo() => new(
        durationSeconds: 120,
        availableQualities: new List<QualityInfo>
        {
            new("720p", 1280, 720, 2800, 12)
        },
        manifestPath: "test-id/manifest.m3u8",
        thumbnailPath: "test-id/thumb_300x170.jpg",
        posterPath: "test-id/poster.jpg",
        encodedAt: DateTime.UtcNow
    );

    [Fact]
    public async Task Handle_MovieNotFound_ReturnsFalse()
    {
        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync((Movie?)null);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = "nonexistent-id",
            VideoStatus = VideoStatus.Uploading
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.False(result);
        _movieRepositoryMock.Verify(
            r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task Handle_ValidTransition_EncodingToReady_ReturnsTrue()
    {
        var movie = CreateMovie(VideoStatus.Encoding);
        var streamingInfo = CreateStreamingInfo();

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);
        _movieRepositoryMock
            .Setup(r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = streamingInfo
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.True(result);
        _movieRepositoryMock.Verify(
            r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()),
            Times.Once);
    }

    [Fact]
    public async Task Handle_InvalidTransition_NotUploadedToReady_ReturnsFalse()
    {
        var movie = CreateMovie(VideoStatus.NotUploaded);

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = CreateStreamingInfo()
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.False(result);
        _movieRepositoryMock.Verify(
            r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()),
            Times.Never);
    }

    [Fact]
    public async Task Handle_ValidTransition_SetsMovieProperties()
    {
        var movie = CreateMovie(VideoStatus.Encoding);
        var streamingInfo = CreateStreamingInfo();
        var beforeTest = DateTime.UtcNow;

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);
        _movieRepositoryMock
            .Setup(r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = streamingInfo
        };

        await _handler.Handle(command, CancellationToken.None);

        Assert.Equal(VideoStatus.Ready, movie.VideoStatus);
        Assert.NotNull(movie.StreamingInfo);
        Assert.Equal(120, movie.StreamingInfo!.DurationSeconds);
        Assert.True(movie.UpdatedAt >= beforeTest);
    }

    [Fact]
    public async Task Handle_ErrorTransition_ClearsStreamingInfo()
    {
        var movie = CreateMovie(VideoStatus.Encoding);
        movie.StreamingInfo = CreateStreamingInfo();

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);
        _movieRepositoryMock
            .Setup(r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = VideoStatus.Error,
            StreamingInfo = null
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.True(result);
        Assert.Equal(VideoStatus.Error, movie.VideoStatus);
        Assert.Null(movie.StreamingInfo);
    }

    [Theory]
    [InlineData(VideoStatus.NotUploaded, VideoStatus.Uploading)]
    [InlineData(VideoStatus.Uploading, VideoStatus.Queued)]
    [InlineData(VideoStatus.Uploading, VideoStatus.Error)]
    [InlineData(VideoStatus.Queued, VideoStatus.Encoding)]
    [InlineData(VideoStatus.Queued, VideoStatus.Error)]
    [InlineData(VideoStatus.Encoding, VideoStatus.Ready)]
    [InlineData(VideoStatus.Encoding, VideoStatus.Error)]
    [InlineData(VideoStatus.Ready, VideoStatus.Uploading)]
    [InlineData(VideoStatus.Error, VideoStatus.Uploading)]
    public async Task Handle_AllValidTransitions_ReturnsTrue(VideoStatus from, VideoStatus to)
    {
        var movie = CreateMovie(from);

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);
        _movieRepositoryMock
            .Setup(r => r.UpdateAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = to,
            StreamingInfo = to == VideoStatus.Ready ? CreateStreamingInfo() : null
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.True(result);
    }

    [Theory]
    [InlineData(VideoStatus.NotUploaded, VideoStatus.Ready)]
    [InlineData(VideoStatus.NotUploaded, VideoStatus.Encoding)]
    [InlineData(VideoStatus.Uploading, VideoStatus.Ready)]
    [InlineData(VideoStatus.Ready, VideoStatus.Encoding)]
    [InlineData(VideoStatus.Queued, VideoStatus.Ready)]
    public async Task Handle_InvalidTransitions_ReturnsFalse(VideoStatus from, VideoStatus to)
    {
        var movie = CreateMovie(from);

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);

        var command = new UpdateVideoStatusCommand
        {
            ContentId = movie.Id,
            VideoStatus = to,
            StreamingInfo = to == VideoStatus.Ready ? CreateStreamingInfo() : null
        };

        var result = await _handler.Handle(command, CancellationToken.None);

        Assert.False(result);
    }

    [Fact]
    public void Constructor_NullRepository_ThrowsArgumentNullException()
    {
        Assert.Throws<ArgumentNullException>(() =>
            new UpdateVideoStatusHandler(null!, _loggerMock.Object));
    }

    [Fact]
    public void Constructor_NullLogger_ThrowsArgumentNullException()
    {
        Assert.Throws<ArgumentNullException>(() =>
            new UpdateVideoStatusHandler(_movieRepositoryMock.Object, null!));
    }
}
