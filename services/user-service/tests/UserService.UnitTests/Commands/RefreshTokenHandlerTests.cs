using Moq;
using UserService.Application.Commands.RefreshToken;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Enums;
using Xunit;

namespace UserService.UnitTests.Commands;

public class RefreshTokenHandlerTests
{
    [Fact]
    public async Task Handle_ShouldReportConfiguredAccessTokenLifetime()
    {
        var userId = Guid.NewGuid();
        var tokenService = new Mock<ITokenService>();
        var userRepository = new Mock<IUserRepository>();

        tokenService.Setup(t => t.ConsumeRefreshTokenAsync("old-refresh", It.IsAny<CancellationToken>()))
            .ReturnsAsync(userId);
        tokenService.Setup(t => t.GenerateAccessToken(It.IsAny<User>())).Returns("access-token");
        tokenService.Setup(t => t.GenerateRefreshTokenAsync(It.IsAny<User>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync("new-refresh");
        tokenService.SetupGet(t => t.AccessTokenExpirationSeconds).Returns(17 * 60);
        userRepository.Setup(r => r.GetSummaryByIdAsync(userId, It.IsAny<CancellationToken>()))
            .ReturnsAsync(new UserReadModel
            {
                Id = userId,
                Email = "refresh@example.com",
                PasswordHash = "hash",
                Role = SubscriptionTier.Free,
                ProfileCount = 1
            });

        var result = await new RefreshTokenHandler(tokenService.Object, userRepository.Object)
            .Handle(new RefreshTokenCommand("old-refresh"), CancellationToken.None);

        Assert.Equal(17 * 60, result.ExpiresIn);
        Assert.Equal("access-token", result.AccessToken);
        Assert.Equal("new-refresh", result.RefreshToken);
    }
}
