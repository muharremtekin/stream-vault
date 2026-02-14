using FluentValidation;

namespace SubscriptionService.Application.Commands.CreateSubscription;

public class CreateSubscriptionValidator : AbstractValidator<CreateSubscriptionCommand>
{
    public CreateSubscriptionValidator()
    {
        RuleFor(x => x.UserId)
            .NotEmpty().WithMessage("UserId is required.");

        RuleFor(x => x.PlanId)
            .NotEmpty().WithMessage("PlanId is required.");

        RuleFor(x => x.CardNumber)
            .NotEmpty().WithMessage("Card number is required.")
            .CreditCard().WithMessage("A valid card number is required.");
    }
}
