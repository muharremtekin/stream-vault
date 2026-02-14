using FluentValidation;

namespace UserService.Application.Commands.RateContent;

public class RateContentValidator : AbstractValidator<RateContentCommand>
{
    public RateContentValidator()
    {
        RuleFor(x => x.ContentId)
            .NotEmpty().WithMessage("ContentId is required.")
            .MaximumLength(256).WithMessage("ContentId must not exceed 256 characters.");

        RuleFor(x => x.Rating)
            .InclusiveBetween(1, 10).WithMessage("Rating must be between 1 and 10.");
    }
}
