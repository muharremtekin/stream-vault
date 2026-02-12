using CatalogService.Application.Interfaces;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using MediatR;

namespace CatalogService.Application.Commands.CreateMovie;

public class CreateMovieHandler : IRequestHandler<CreateMovieCommand, string>
{
    private readonly IMovieRepository _movieRepository;

    public CreateMovieHandler(IMovieRepository movieRepository)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
    }

    public async Task<string> Handle(CreateMovieCommand request, CancellationToken cancellationToken)
    {
        var movie = new Movie
        {
            Title = request.Title,
            OriginalTitle = request.OriginalTitle,
            Description = request.Description,
            ReleaseYear = request.ReleaseYear,
            Duration = Duration.FromMinutes(request.DurationMinutes),
            MaturityRating = request.MaturityRating,
            Genres = request.Genres,
            Cast = request.Cast.Select(c => new CastMember(c.Name, c.Role, c.PhotoUrl)).ToList(),
            Director = request.Director,
            ThumbnailUrl = request.ThumbnailUrl,
            BannerUrl = request.BannerUrl,
            TrailerUrl = request.TrailerUrl,
            Tags = request.Tags,
            Status = ContentStatus.Draft,
            AverageRating = 0,
            RatingCount = 0,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };

        await _movieRepository.AddAsync(movie, cancellationToken);

        return movie.Id;
    }
}
