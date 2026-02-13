using CatalogService.Domain.Enums;
using FluentValidation;

namespace CatalogService.Application.Commands.UpdateVideoStatus;

public class UpdateVideoStatusValidator : AbstractValidator<UpdateVideoStatusCommand>
{
    public UpdateVideoStatusValidator()
    {
        RuleFor(x => x.ContentId)
            .NotEmpty().WithMessage("Content ID is required.");

        RuleFor(x => x.VideoStatus)
            .IsInEnum().WithMessage("Invalid video status.");

        RuleFor(x => x.StreamingInfo)
            .NotNull()
            .When(x => x.VideoStatus == VideoStatus.Ready)
            .WithMessage("Streaming info is required when video status is Ready.");

        RuleFor(x => x.StreamingInfo)
            .Null()
            .When(x => x.VideoStatus == VideoStatus.Error)
            .WithMessage("Streaming info must be null when video status is Error.");
    }
}
