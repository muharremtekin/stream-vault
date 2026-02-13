using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using UserService.Application.Commands.LoginUser;
using UserService.Application.Commands.RegisterUser;
using UserService.Application.DTOs;
using Xunit;

namespace UserService.IntegrationTests;

public class AuthControllerTests : IClassFixture<CustomWebApplicationFactory>
{
    private readonly HttpClient _client;
    private readonly CustomWebApplicationFactory _factory;

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase
    };

    public AuthControllerTests(CustomWebApplicationFactory factory)
    {
        _factory = factory;
        _client = factory.CreateClient();
    }

    [Fact]
    public async Task Register_WithValidData_ShouldReturn201Created()
    {
        // Arrange
        var command = new RegisterUserCommand(
            $"test_{Guid.NewGuid():N}@example.com",
            "P@ssword123!");

        // Act
        var response = await _client.PostAsJsonAsync("/api/auth/register", command);

        // Assert
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);

        var content = await response.Content.ReadFromJsonAsync<AuthResponseDto>(JsonOptions);
        Assert.NotNull(content);
        Assert.NotEmpty(content.AccessToken);
        Assert.NotEmpty(content.RefreshToken);
        Assert.NotNull(content.User);
        Assert.Equal(command.Email, content.User.Email);
    }

    [Fact]
    public async Task Register_WithDuplicateEmail_ShouldReturn409Conflict()
    {
        // Arrange
        var email = $"duplicate_{Guid.NewGuid():N}@example.com";
        var command = new RegisterUserCommand(email, "P@ssword123!");

        // Register the first user
        await _client.PostAsJsonAsync("/api/auth/register", command);

        // Act - try to register with the same email
        var response = await _client.PostAsJsonAsync("/api/auth/register", command);

        // Assert
        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }

    [Fact]
    public async Task Login_WithValidCredentials_ShouldReturn200Ok()
    {
        // Arrange
        var email = $"login_{Guid.NewGuid():N}@example.com";
        var password = "P@ssword123!";

        var registerCommand = new RegisterUserCommand(email, password);
        await _client.PostAsJsonAsync("/api/auth/register", registerCommand);

        var loginCommand = new LoginUserCommand(email, password);

        // Act
        var response = await _client.PostAsJsonAsync("/api/auth/login", loginCommand);

        // Assert
        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        var content = await response.Content.ReadFromJsonAsync<AuthResponseDto>(JsonOptions);
        Assert.NotNull(content);
        Assert.NotEmpty(content.AccessToken);
        Assert.Equal(email, content.User.Email);
    }

    [Fact]
    public async Task Login_WithInvalidPassword_ShouldReturn404NotFound()
    {
        // Arrange
        var email = $"wrongpass_{Guid.NewGuid():N}@example.com";

        var registerCommand = new RegisterUserCommand(email, "P@ssword123!");
        await _client.PostAsJsonAsync("/api/auth/register", registerCommand);

        var loginCommand = new LoginUserCommand(email, "WrongP@ssword!");

        // Act
        var response = await _client.PostAsJsonAsync("/api/auth/login", loginCommand);

        // Assert
        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task HealthCheck_ShouldReturn200Ok()
    {
        // Act
        var response = await _client.GetAsync("/health");

        // Assert
        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task Register_WithInvalidPassword_ShouldReturn400BadRequest()
    {
        // Arrange — weak password: too short, no uppercase, no digit, no special char
        var command = new RegisterUserCommand(
            $"weakpass_{Guid.NewGuid():N}@example.com",
            "weak");

        // Act
        var response = await _client.PostAsJsonAsync("/api/auth/register", command);

        // Assert
        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task Register_ThenLogin_FullFlow_ShouldReturnConsistentUserData()
    {
        // Arrange
        var email = $"flow_{Guid.NewGuid():N}@example.com";
        var password = "P@ssword123!";

        // Act — register
        var registerResponse = await _client.PostAsJsonAsync(
            "/api/auth/register",
            new RegisterUserCommand(email, password));

        Assert.Equal(HttpStatusCode.Created, registerResponse.StatusCode);

        var registerContent = await registerResponse.Content
            .ReadFromJsonAsync<AuthResponseDto>(JsonOptions);
        Assert.NotNull(registerContent);

        // Act — login
        var loginResponse = await _client.PostAsJsonAsync(
            "/api/auth/login",
            new LoginUserCommand(email, password));

        Assert.Equal(HttpStatusCode.OK, loginResponse.StatusCode);

        var loginContent = await loginResponse.Content
            .ReadFromJsonAsync<AuthResponseDto>(JsonOptions);
        Assert.NotNull(loginContent);

        // Assert — consistent user data
        Assert.Equal(registerContent.User.Email, loginContent.User.Email);
        Assert.Equal(registerContent.User.Id, loginContent.User.Id);
        Assert.NotEmpty(loginContent.AccessToken);
        Assert.NotEmpty(loginContent.RefreshToken);
    }
}
