using CatalogService.Application.Interfaces;
using CatalogService.Application.Queries.GetStreamingInfo;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using Moq;
using Xunit;

namespace CatalogService.UnitTests.Queries;

public class GetStreamingInfoHandlerTests
{
    private readonly Mock<IMovieRepository> _movieRepositoryMock;
    private readonly GetStreamingInfoHandler _handler;

    public GetStreamingInfoHandlerTests()
    {
        _movieRepositoryMock = new Mock<IMovieRepository>();
        _handler = new GetStreamingInfoHandler(_movieRepositoryMock.Object);
    }

    [Fact]
    public async Task Handle_MovieNotFound_ReturnsNull()
    {
        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync((Movie?)null);

        var query = new GetStreamingInfoQuery { MovieId = "nonexistent-id" };

        var result = await _handler.Handle(query, CancellationToken.None);

        Assert.Null(result);
    }

    [Fact]
    public async Task Handle_MovieWithoutStreamingInfo_ReturnsVideoStatusOnly()
    {
        var movie = new Movie
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
            VideoStatus = VideoStatus.NotUploaded,
            StreamingInfo = null
        };

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);

        var query = new GetStreamingInfoQuery { MovieId = movie.Id };

        var result = await _handler.Handle(query, CancellationToken.None);

        Assert.NotNull(result);
        Assert.Equal("NotUploaded", result!.VideoStatus);
        Assert.Null(result.DurationSeconds);
        Assert.Empty(result.AvailableQualities);
        Assert.Null(result.ManifestUrl);
        Assert.Null(result.ThumbnailUrl);
        Assert.Null(result.PosterUrl);
        Assert.Null(result.EncodedAt);
    }

    [Fact]
    public async Task Handle_MovieWithStreamingInfo_ReturnsFullDto()
    {
        var encodedAt = new DateTime(2024, 6, 15, 10, 30, 0, DateTimeKind.Utc);
        var movie = new Movie
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
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = new StreamingInfo(
                durationSeconds: 120,
                availableQualities: new List<QualityInfo>
                {
                    new("360p", 640, 360, 800, 12),
                    new("720p", 1280, 720, 2800, 12)
                },
                manifestPath: "507f1f77bcf86cd799439011/manifest.m3u8",
                thumbnailPath: "507f1f77bcf86cd799439011/thumb_300x170.jpg",
                posterPath: "507f1f77bcf86cd799439011/poster.jpg",
                encodedAt: encodedAt
            )
        };

        _movieRepositoryMock
            .Setup(r => r.GetByIdAsync(movie.Id, It.IsAny<CancellationToken>()))
            .ReturnsAsync(movie);

        var query = new GetStreamingInfoQuery { MovieId = movie.Id };

        var result = await _handler.Handle(query, CancellationToken.None);

        Assert.NotNull(result);
        Assert.Equal("Ready", result!.VideoStatus);
        Assert.Equal(120, result.DurationSeconds);
        Assert.Equal(2, result.AvailableQualities.Count);

        Assert.Equal("360p", result.AvailableQualities[0].Label);
        Assert.Equal(640, result.AvailableQualities[0].Width);
        Assert.Equal(360, result.AvailableQualities[0].Height);
        Assert.Equal(800, result.AvailableQualities[0].BitrateKbps);
        Assert.Equal(12, result.AvailableQualities[0].SegmentCount);

        Assert.Equal("720p", result.AvailableQualities[1].Label);
        Assert.Equal(1280, result.AvailableQualities[1].Width);

        Assert.Equal("507f1f77bcf86cd799439011/manifest.m3u8", result.ManifestUrl);
        Assert.Equal("507f1f77bcf86cd799439011/thumb_300x170.jpg", result.ThumbnailUrl);
        Assert.Equal("507f1f77bcf86cd799439011/poster.jpg", result.PosterUrl);
        Assert.Equal(encodedAt, result.EncodedAt);
    }

    [Fact]
    public void Constructor_NullRepository_ThrowsArgumentNullException()
    {
        Assert.Throws<ArgumentNullException>(() => new GetStreamingInfoHandler(null!));
    }
}
