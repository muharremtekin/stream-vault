using FluentValidation;

namespace SubscriptionService.Application.Commands.ChangePlan;

public class ChangePlanValidator : AbstractValidator<ChangePlanCommand>
{
    public ChangePlanValidator()
    {
        RuleFor(x => x.UserId)
            .NotEmpty().WithMessage("UserId is required.");

        RuleFor(x => x.NewPlanId)
            .NotEmpty().WithMessage("NewPlanId is required.");

        RuleFor(x => x.CardNumber)
            .NotEmpty().WithMessage("Card number is required.")
            .CreditCard().WithMessage("A valid card number is required.");
    }
}
