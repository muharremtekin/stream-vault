using Moq;
using UserService.Application.Commands.LoginUser;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Enums;
using UserService.Domain.Exceptions;
using Xunit;

namespace UserService.UnitTests.Commands;

public class LoginUserHandlerTests
{
    private readonly Mock<IUserRepository> _userRepositoryMock;
    private readonly Mock<IPasswordHasher> _passwordHasherMock;
    private readonly Mock<ITokenService> _tokenServiceMock;
    private readonly LoginUserHandler _handler;

    public LoginUserHandlerTests()
    {
        _userRepositoryMock = new Mock<IUserRepository>();
        _passwordHasherMock = new Mock<IPasswordHasher>();
        _tokenServiceMock = new Mock<ITokenService>();

        _handler = new LoginUserHandler(
            _userRepositoryMock.Object,
            _passwordHasherMock.Object,
            _tokenServiceMock.Object);
    }

    [Fact]
    public async Task Handle_WithValidCredentials_ShouldReturnAuthResponse()
    {
        // Arrange
        var user = new UserReadModel
        {
            Id = Guid.NewGuid(),
            Email = "test@example.com",
            PasswordHash = "hashed_password",
            Role = SubscriptionTier.Free,
            ProfileCount = 0
        };

        var command = new LoginUserCommand("test@example.com", "P@ssword123!");

        _userRepositoryMock
            .Setup(r => r.GetSummaryByEmailAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync(user);

        _passwordHasherMock
            .Setup(h => h.Verify(command.Password, user.PasswordHash))
            .Returns(true);

        _tokenServiceMock
            .Setup(t => t.GenerateAccessToken(
                It.Is<User>(u =>
                    u.Id == user.Id &&
                    u.Email == user.Email &&
                    u.PasswordHash == user.PasswordHash &&
                    u.Role == user.Role)))
            .Returns("access_token");

        _tokenServiceMock
            .Setup(t => t.GenerateRefreshTokenAsync(
                It.Is<User>(u =>
                    u.Id == user.Id &&
                    u.Email == user.Email &&
                    u.PasswordHash == user.PasswordHash &&
                    u.Role == user.Role),
                It.IsAny<CancellationToken>()))
            .ReturnsAsync("refresh_token");

        _tokenServiceMock
            .SetupGet(t => t.AccessTokenExpirationSeconds)
            .Returns(17 * 60);

        // Act
        var result = await _handler.Handle(command, CancellationToken.None);

        // Assert
        Assert.NotNull(result);
        Assert.Equal("access_token", result.AccessToken);
        Assert.Equal("refresh_token", result.RefreshToken);
        Assert.Equal(17 * 60, result.ExpiresIn);
        Assert.Equal(user.Email, result.User.Email);
    }

    [Fact]
    public async Task Handle_WithNonExistentEmail_ShouldThrowUserNotFoundException()
    {
        // Arrange
        var command = new LoginUserCommand("nonexistent@example.com", "P@ssword123!");

        _userRepositoryMock
            .Setup(r => r.GetSummaryByEmailAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync((UserReadModel?)null);

        // Act & Assert
        await Assert.ThrowsAsync<UserNotFoundException>(
            () => _handler.Handle(command, CancellationToken.None));
    }

    [Fact]
    public async Task Handle_WithInvalidPassword_ShouldThrowUserNotFoundException()
    {
        // Arrange
        var user = new UserReadModel
        {
            Id = Guid.NewGuid(),
            Email = "test@example.com",
            PasswordHash = "hashed_password",
            Role = SubscriptionTier.Free
        };

        var command = new LoginUserCommand("test@example.com", "WrongPassword!");

        _userRepositoryMock
            .Setup(r => r.GetSummaryByEmailAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync(user);

        _passwordHasherMock
            .Setup(h => h.Verify(command.Password, user.PasswordHash))
            .Returns(false);

        // Act & Assert
        await Assert.ThrowsAsync<UserNotFoundException>(
            () => _handler.Handle(command, CancellationToken.None));
    }
}
