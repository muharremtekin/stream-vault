using Moq;
using UserService.Application.Commands.RegisterUser;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Exceptions;
using Xunit;

namespace UserService.UnitTests.Commands;

public class RegisterUserHandlerTests
{
    private readonly Mock<IUserRepository> _userRepositoryMock;
    private readonly Mock<IPasswordHasher> _passwordHasherMock;
    private readonly Mock<ITokenService> _tokenServiceMock;
    private readonly RegisterUserHandler _handler;

    public RegisterUserHandlerTests()
    {
        _userRepositoryMock = new Mock<IUserRepository>();
        _passwordHasherMock = new Mock<IPasswordHasher>();
        _tokenServiceMock = new Mock<ITokenService>();

        _handler = new RegisterUserHandler(
            _userRepositoryMock.Object,
            _passwordHasherMock.Object,
            _tokenServiceMock.Object);
    }

    [Fact]
    public async Task Handle_WithValidCommand_ShouldReturnAuthResponse()
    {
        // Arrange
        var command = new RegisterUserCommand("test@example.com", "P@ssword123!");

        _userRepositoryMock
            .Setup(r => r.ExistsAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync(false);

        _passwordHasherMock
            .Setup(h => h.Hash(command.Password))
            .Returns("hashed_password");

        _tokenServiceMock
            .Setup(t => t.GenerateAccessToken(It.IsAny<User>()))
            .Returns("access_token");

        _tokenServiceMock
            .Setup(t => t.GenerateRefreshTokenAsync(It.IsAny<User>(), It.IsAny<CancellationToken>()))
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
        Assert.Equal("test@example.com", result.User.Email);
        _userRepositoryMock.Verify(r => r.AddAsync(It.IsAny<User>(), It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task Handle_WithDuplicateEmail_ShouldThrowDuplicateEmailException()
    {
        // Arrange
        var command = new RegisterUserCommand("existing@example.com", "P@ssword123!");

        _userRepositoryMock
            .Setup(r => r.ExistsAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync(true);

        // Act & Assert
        await Assert.ThrowsAsync<DuplicateEmailException>(
            () => _handler.Handle(command, CancellationToken.None));
    }

    [Fact]
    public async Task Handle_ShouldHashPasswordBeforeStoring()
    {
        // Arrange
        var command = new RegisterUserCommand("test@example.com", "P@ssword123!");

        _userRepositoryMock
            .Setup(r => r.ExistsAsync(command.Email, It.IsAny<CancellationToken>()))
            .ReturnsAsync(false);

        _passwordHasherMock
            .Setup(h => h.Hash(command.Password))
            .Returns("hashed_password");

        _tokenServiceMock
            .Setup(t => t.GenerateAccessToken(It.IsAny<User>()))
            .Returns("token");

        _tokenServiceMock
            .Setup(t => t.GenerateRefreshTokenAsync(It.IsAny<User>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync("refresh");

        // Act
        await _handler.Handle(command, CancellationToken.None);

        // Assert
        _passwordHasherMock.Verify(h => h.Hash(command.Password), Times.Once);
        _userRepositoryMock.Verify(r => r.AddAsync(
            It.Is<User>(u => u.PasswordHash == "hashed_password"),
            It.IsAny<CancellationToken>()), Times.Once);
    }
}
