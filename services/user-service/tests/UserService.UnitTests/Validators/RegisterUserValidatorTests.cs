using FluentValidation.TestHelper;
using UserService.Application.Commands.RegisterUser;
using Xunit;

namespace UserService.UnitTests.Validators;

public class RegisterUserValidatorTests
{
    private readonly RegisterUserValidator _validator = new();

    [Fact]
    public void Validate_EmptyEmail_ShouldHaveError()
    {
        var command = new RegisterUserCommand("", "P@ssword123!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Email)
            .WithErrorMessage("Email is required.");
    }

    [Fact]
    public void Validate_InvalidEmailFormat_ShouldHaveError()
    {
        var command = new RegisterUserCommand("not-an-email", "P@ssword123!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Email)
            .WithErrorMessage("A valid email address is required.");
    }

    [Fact]
    public void Validate_EmptyPassword_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password is required.");
    }

    [Fact]
    public void Validate_ShortPassword_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "Ab1!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password must be at least 8 characters.");
    }

    [Fact]
    public void Validate_PasswordWithoutUppercase_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "p@ssword123!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password must contain at least one uppercase letter.");
    }

    [Fact]
    public void Validate_PasswordWithoutLowercase_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "P@SSWORD123!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password must contain at least one lowercase letter.");
    }

    [Fact]
    public void Validate_PasswordWithoutDigit_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "P@sswordABC!");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password must contain at least one digit.");
    }

    [Fact]
    public void Validate_PasswordWithoutSpecialChar_ShouldHaveError()
    {
        var command = new RegisterUserCommand("test@example.com", "Password123a");
        var result = _validator.TestValidate(command);
        result.ShouldHaveValidationErrorFor(x => x.Password)
            .WithErrorMessage("Password must contain at least one special character.");
    }

    [Fact]
    public void Validate_ValidCommand_ShouldNotHaveErrors()
    {
        var command = new RegisterUserCommand("test@example.com", "P@ssword123!");
        var result = _validator.TestValidate(command);
        result.ShouldNotHaveAnyValidationErrors();
    }
}
