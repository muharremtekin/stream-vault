using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Domain.Enums;
using FluentValidation.TestHelper;
using Xunit;

namespace CatalogService.UnitTests.Validators;

public class CreateMovieValidatorTests
{
    private readonly CreateMovieValidator _validator = new();

    private static CreateMovieCommand CreateValidCommand() => new()
    {
        Title = "Test Movie",
        Description = "A valid test movie description.",
        ReleaseYear = 2024,
        DurationMinutes = 120,
        MaturityRating = MaturityRating.PG13,
        Genres = new List<string> { "Action" },
        Cast = new List<CastMemberInput> { new("Actor", "Lead") },
        Director = "Test Director",
        ThumbnailUrl = "https://cdn.example.com/thumb.jpg",
        BannerUrl = "https://cdn.example.com/banner.jpg",
        Tags = new List<string> { "test" }
    };

    [Fact]
    public void Validate_EmptyTitle_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { Title = "" };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Title)
            .WithErrorMessage("Title is required.");
    }

    [Fact]
    public void Validate_TitleExceedsMaxLength_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { Title = new string('A', 201) };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Title)
            .WithErrorMessage("Title must not exceed 200 characters.");
    }

    [Fact]
    public void Validate_EmptyDescription_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { Description = "" };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Description)
            .WithErrorMessage("Description is required.");
    }

    [Fact]
    public void Validate_ReleaseYearTooLow_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { ReleaseYear = 1887 };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.ReleaseYear);
    }

    [Fact]
    public void Validate_ReleaseYearTooHigh_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { ReleaseYear = DateTime.UtcNow.Year + 6 };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.ReleaseYear);
    }

    [Fact]
    public void Validate_DurationZero_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { DurationMinutes = 0 };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.DurationMinutes)
            .WithErrorMessage("Duration must be greater than 0 minutes.");
    }

    [Fact]
    public void Validate_DurationExceedsMax_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { DurationMinutes = 601 };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.DurationMinutes)
            .WithErrorMessage("Duration must not exceed 600 minutes.");
    }

    [Fact]
    public void Validate_EmptyGenres_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { Genres = new List<string>() };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Genres)
            .WithErrorMessage("At least one genre is required.");
    }

    [Fact]
    public void Validate_EmptyDirector_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { Director = "" };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Director)
            .WithErrorMessage("Director is required.");
    }

    [Fact]
    public void Validate_InvalidThumbnailUrl_ShouldHaveError()
    {
        var command = CreateValidCommand();
        command = command with { ThumbnailUrl = "not-a-url" };
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.ThumbnailUrl)
            .WithErrorMessage("Thumbnail URL must be a valid URL.");
    }

    [Fact]
    public void Validate_ValidCommand_ShouldNotHaveErrors()
    {
        var command = CreateValidCommand();
        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }
}
