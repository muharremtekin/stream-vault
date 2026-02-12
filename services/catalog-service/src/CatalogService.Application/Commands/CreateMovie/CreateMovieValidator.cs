using FluentValidation;

namespace CatalogService.Application.Commands.CreateMovie;

public class CreateMovieValidator : AbstractValidator<CreateMovieCommand>
{
    public CreateMovieValidator()
    {
        RuleFor(x => x.Title)
            .NotEmpty().WithMessage("Title is required.")
            .MaximumLength(200).WithMessage("Title must not exceed 200 characters.");

        RuleFor(x => x.Description)
            .NotEmpty().WithMessage("Description is required.")
            .MaximumLength(2000).WithMessage("Description must not exceed 2000 characters.");

        RuleFor(x => x.ReleaseYear)
            .InclusiveBetween(1888, DateTime.UtcNow.Year + 5)
            .WithMessage("Release year must be between 1888 and 5 years from now.");

        RuleFor(x => x.DurationMinutes)
            .GreaterThan(0).WithMessage("Duration must be greater than 0 minutes.")
            .LessThanOrEqualTo(600).WithMessage("Duration must not exceed 600 minutes.");

        RuleFor(x => x.MaturityRating)
            .IsInEnum().WithMessage("Invalid maturity rating.");

        RuleFor(x => x.Genres)
            .NotEmpty().WithMessage("At least one genre is required.");

        RuleFor(x => x.Director)
            .NotEmpty().WithMessage("Director is required.")
            .MaximumLength(100).WithMessage("Director name must not exceed 100 characters.");

        RuleFor(x => x.ThumbnailUrl)
            .NotEmpty().WithMessage("Thumbnail URL is required.")
            .Must(BeAValidUrl).WithMessage("Thumbnail URL must be a valid URL.");

        RuleFor(x => x.BannerUrl)
            .NotEmpty().WithMessage("Banner URL is required.")
            .Must(BeAValidUrl).WithMessage("Banner URL must be a valid URL.");

        RuleFor(x => x.TrailerUrl)
            .Must(BeAValidUrl)
            .When(x => !string.IsNullOrEmpty(x.TrailerUrl))
            .WithMessage("Trailer URL must be a valid URL.");

        RuleForEach(x => x.Cast).ChildRules(cast =>
        {
            cast.RuleFor(c => c.Name).NotEmpty().WithMessage("Cast member name is required.");
            cast.RuleFor(c => c.Role).NotEmpty().WithMessage("Cast member role is required.");
        });
    }

    private static bool BeAValidUrl(string? url)
    {
        if (string.IsNullOrEmpty(url)) return true;
        return Uri.TryCreate(url, UriKind.Absolute, out var result)
               && (result.Scheme == Uri.UriSchemeHttp || result.Scheme == Uri.UriSchemeHttps);
    }
}
