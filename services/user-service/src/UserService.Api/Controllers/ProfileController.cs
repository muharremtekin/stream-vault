using MediatR;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using UserService.Application.Commands.CreateProfile;
using UserService.Application.DTOs;
using UserService.Application.Queries.GetUserProfiles;
using UserService.Domain.Enums;

namespace UserService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class ProfileController : ControllerBase
{
    private readonly IMediator _mediator;

    public ProfileController(IMediator mediator)
    {
        _mediator = mediator;
    }

    /// <summary>
    /// Get all profiles for a user.
    /// </summary>
    [HttpGet("user/{userId:guid}")]
    [ProducesResponseType(typeof(List<ProfileDto>), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetUserProfiles(
        Guid userId,
        CancellationToken cancellationToken)
    {
        var query = new GetUserProfilesQuery(userId);
        var result = await _mediator.Send(query, cancellationToken);
        return Ok(result);
    }

    /// <summary>
    /// Create a new profile for a user.
    /// </summary>
    [HttpPost]
    [ProducesResponseType(typeof(ProfileDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> CreateProfile(
        [FromBody] CreateProfileRequest request,
        CancellationToken cancellationToken)
    {
        var command = new CreateProfileCommand(
            request.UserId,
            request.Name,
            request.Icon,
            request.IsKids);

        var result = await _mediator.Send(command, cancellationToken);
        return CreatedAtAction(nameof(GetUserProfiles), new { userId = request.UserId }, result);
    }
}

public record CreateProfileRequest(
    Guid UserId,
    string Name,
    ProfileIcon Icon,
    bool IsKids);
