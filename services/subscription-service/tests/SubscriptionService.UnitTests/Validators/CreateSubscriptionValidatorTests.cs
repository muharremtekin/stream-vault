using FluentAssertions;
using FluentValidation.TestHelper;
using Xunit;
using SubscriptionService.Application.Commands.CreateSubscription;

namespace SubscriptionService.UnitTests.Validators;

public class CreateSubscriptionValidatorTests
{
    private readonly CreateSubscriptionValidator _validator = new();

    [Fact]
    public void Validate_ValidCommand_IsValid()
    {
        var command = new CreateSubscriptionCommand(
            Guid.NewGuid(),
            Guid.NewGuid(),
            "4242424242424242");

        var result = _validator.TestValidate(command);

        result.ShouldNotHaveAnyValidationErrors();
    }

    [Fact]
    public void Validate_EmptyUserId_HasValidationError()
    {
        var command = new CreateSubscriptionCommand(
            Guid.Empty,
            Guid.NewGuid(),
            "4242424242424242");

        var result = _validator.TestValidate(command);

        result.ShouldHaveValidationErrorFor(x => x.UserId);
    }

    [Fact]
    public void Validate_EmptyPlanId_HasValidationError()
    {
        var command = new CreateSubscriptionCommand(
            Guid.NewGuid(),
            Guid.Empty,
            "4242424242424242");

        var result = _validator.TestValidate(command);

        result.ShouldHaveValidationErrorFor(x => x.PlanId);
    }

    [Fact]
    public void Validate_InvalidCardNumber_HasValidationError()
    {
        var command = new CreateSubscriptionCommand(
            Guid.NewGuid(),
            Guid.NewGuid(),
            "1234");

        var result = _validator.TestValidate(command);

        result.ShouldHaveValidationErrorFor(x => x.CardNumber);
    }
}
