using CatalogService.Application.Commands.UpdateVideoStatus;
using CatalogService.Domain.Enums;
using CatalogService.Domain.ValueObjects;
using FluentValidation.TestHelper;
using Xunit;

namespace CatalogService.UnitTests.Validators;

public class UpdateVideoStatusValidatorTests
{
    private readonly UpdateVideoStatusValidator _validator = new();

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
    public void Validate_EmptyContentId_ShouldHaveError()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "",
            VideoStatus = VideoStatus.Uploading
        };

        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.ContentId)
            .WithErrorMessage("Content ID is required.");
    }

    [Fact]
    public void Validate_ValidCommand_ShouldNotHaveErrors()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Uploading
        };

        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }

    [Fact]
    public void Validate_ReadyWithoutStreamingInfo_ShouldHaveError()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = null
        };

        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.StreamingInfo)
            .WithErrorMessage("Streaming info is required when video status is Ready.");
    }

    [Fact]
    public void Validate_ReadyWithStreamingInfo_ShouldNotHaveErrors()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Ready,
            StreamingInfo = CreateStreamingInfo()
        };

        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }

    [Fact]
    public void Validate_ErrorWithStreamingInfo_ShouldHaveError()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Error,
            StreamingInfo = CreateStreamingInfo()
        };

        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.StreamingInfo)
            .WithErrorMessage("Streaming info must be null when video status is Error.");
    }

    [Fact]
    public void Validate_ErrorWithoutStreamingInfo_ShouldNotHaveErrors()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Error,
            StreamingInfo = null
        };

        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }

    [Fact]
    public void Validate_EncodingWithoutStreamingInfo_ShouldNotHaveErrors()
    {
        var command = new UpdateVideoStatusCommand
        {
            ContentId = "507f1f77bcf86cd799439011",
            VideoStatus = VideoStatus.Encoding,
            StreamingInfo = null
        };

        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }
}
