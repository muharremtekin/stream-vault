using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using Moq;
using Xunit;

namespace CatalogService.UnitTests.Commands;

public class CreateMovieHandlerTests
{
    private readonly Mock<IMovieRepository> _movieRepositoryMock;
    private readonly CreateMovieHandler _handler;

    public CreateMovieHandlerTests()
    {
        _movieRepositoryMock = new Mock<IMovieRepository>();
        _handler = new CreateMovieHandler(_movieRepositoryMock.Object);
    }

    [Fact]
    public async Task Handle_ValidCommand_ReturnsMovieId()
    {
        // Arrange
        _movieRepositoryMock
            .Setup(r => r.AddAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Returns(Task.CompletedTask);

        var command = new CreateMovieCommand
        {
            Title = "Test Movie",
            Description = "A test movie description.",
            ReleaseYear = 2024,
            DurationMinutes = 120,
            MaturityRating = MaturityRating.PG13,
            Genres = new List<string> { "Action", "Drama" },
            Cast = new List<CastMemberInput>
            {
                new("Actor One", "Lead Role"),
                new("Actor Two", "Supporting Role")
            },
            Director = "Test Director",
            ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/test.jpg",
            BannerUrl = "https://cdn.streamvault.io/banners/test.jpg",
            Tags = new List<string> { "test" }
        };

        // Act
        var result = await _handler.Handle(command, CancellationToken.None);

        // Assert
        Assert.NotNull(result);
        Assert.NotEmpty(result);
        _movieRepositoryMock.Verify(
            r => r.AddAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()),
            Times.Once);
    }

    [Fact]
    public async Task Handle_ValidCommand_SetsCorrectMovieProperties()
    {
        // Arrange
        Movie? capturedMovie = null;
        _movieRepositoryMock
            .Setup(r => r.AddAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Callback<Movie, CancellationToken>((movie, _) => capturedMovie = movie)
            .Returns(Task.CompletedTask);

        var command = new CreateMovieCommand
        {
            Title = "Inception Test",
            OriginalTitle = "Inception Original",
            Description = "A mind-bending test.",
            ReleaseYear = 2010,
            DurationMinutes = 148,
            MaturityRating = MaturityRating.PG13,
            Genres = new List<string> { "Science Fiction", "Thriller" },
            Cast = new List<CastMemberInput>
            {
                new("Leonardo DiCaprio", "Dom Cobb", "https://cdn.streamvault.io/photos/leo.jpg")
            },
            Director = "Christopher Nolan",
            ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/inception-test.jpg",
            BannerUrl = "https://cdn.streamvault.io/banners/inception-test.jpg",
            TrailerUrl = "https://cdn.streamvault.io/trailers/inception-test.mp4",
            Tags = new List<string> { "dreams", "heist" }
        };

        // Act
        await _handler.Handle(command, CancellationToken.None);

        // Assert
        Assert.NotNull(capturedMovie);
        Assert.Equal("Inception Test", capturedMovie!.Title);
        Assert.Equal("Inception Original", capturedMovie.OriginalTitle);
        Assert.Equal("A mind-bending test.", capturedMovie.Description);
        Assert.Equal(2010, capturedMovie.ReleaseYear);
        Assert.Equal(148, capturedMovie.Duration.TotalMinutes);
        Assert.Equal(MaturityRating.PG13, capturedMovie.MaturityRating);
        Assert.Equal(2, capturedMovie.Genres.Count);
        Assert.Contains("Science Fiction", capturedMovie.Genres);
        Assert.Single(capturedMovie.Cast);
        Assert.Equal("Leonardo DiCaprio", capturedMovie.Cast[0].Name);
        Assert.Equal("Christopher Nolan", capturedMovie.Director);
        Assert.Equal(ContentStatus.Draft, capturedMovie.Status);
        Assert.Equal(0, capturedMovie.AverageRating);
        Assert.Equal(0, capturedMovie.RatingCount);
        Assert.NotNull(capturedMovie.TrailerUrl);
        Assert.Equal(2, capturedMovie.Tags.Count);
    }

    [Fact]
    public async Task Handle_ValidCommand_SetsStatusToDraft()
    {
        // Arrange
        Movie? capturedMovie = null;
        _movieRepositoryMock
            .Setup(r => r.AddAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Callback<Movie, CancellationToken>((movie, _) => capturedMovie = movie)
            .Returns(Task.CompletedTask);

        var command = new CreateMovieCommand
        {
            Title = "Draft Test Movie",
            Description = "Testing draft status.",
            ReleaseYear = 2024,
            DurationMinutes = 90,
            MaturityRating = MaturityRating.G,
            Genres = new List<string> { "Comedy" },
            Cast = new List<CastMemberInput>(),
            Director = "Test Director",
            ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/draft-test.jpg",
            BannerUrl = "https://cdn.streamvault.io/banners/draft-test.jpg"
        };

        // Act
        await _handler.Handle(command, CancellationToken.None);

        // Assert
        Assert.NotNull(capturedMovie);
        Assert.Equal(ContentStatus.Draft, capturedMovie!.Status);
    }

    [Fact]
    public async Task Handle_ValidCommand_SetsTimestamps()
    {
        // Arrange
        Movie? capturedMovie = null;
        var beforeTest = DateTime.UtcNow;

        _movieRepositoryMock
            .Setup(r => r.AddAsync(It.IsAny<Movie>(), It.IsAny<CancellationToken>()))
            .Callback<Movie, CancellationToken>((movie, _) => capturedMovie = movie)
            .Returns(Task.CompletedTask);

        var command = new CreateMovieCommand
        {
            Title = "Timestamp Test",
            Description = "Testing timestamps.",
            ReleaseYear = 2024,
            DurationMinutes = 100,
            MaturityRating = MaturityRating.PG,
            Genres = new List<string> { "Drama" },
            Cast = new List<CastMemberInput>(),
            Director = "Director",
            ThumbnailUrl = "https://cdn.streamvault.io/thumbnails/ts-test.jpg",
            BannerUrl = "https://cdn.streamvault.io/banners/ts-test.jpg"
        };

        // Act
        await _handler.Handle(command, CancellationToken.None);

        // Assert
        Assert.NotNull(capturedMovie);
        Assert.True(capturedMovie!.CreatedAt >= beforeTest);
        Assert.True(capturedMovie.UpdatedAt >= beforeTest);
        Assert.True(capturedMovie.CreatedAt <= DateTime.UtcNow);
    }

    [Fact]
    public void Constructor_NullRepository_ThrowsArgumentNullException()
    {
        // Act & Assert
        Assert.Throws<ArgumentNullException>(() => new CreateMovieHandler(null!));
    }
}
