using MediatR;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using UserService.Application.Commands.AddToWatchlist;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Application.Queries.GetWatchlist;
using UserService.Domain.Enums;

namespace UserService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class WatchlistController : ControllerBase
{
    private readonly IMediator _mediator;
    private readonly IWatchlistRepository _watchlistRepository;

    public WatchlistController(IMediator mediator, IWatchlistRepository watchlistRepository)
    {
        _mediator = mediator;
        _watchlistRepository = watchlistRepository;
    }

    /// <summary>
    /// Get all watchlist items for a profile.
    /// </summary>
    [HttpGet("profile/{profileId:guid}")]
    [ProducesResponseType(typeof(List<WatchlistItemDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetWatchlist(
        Guid profileId,
        CancellationToken cancellationToken)
    {
        var query = new GetWatchlistQuery(profileId);
        var result = await _mediator.Send(query, cancellationToken);
        return Ok(result);
    }

    /// <summary>
    /// Add an item to the watchlist.
    /// </summary>
    [HttpPost]
    [ProducesResponseType(typeof(WatchlistItemDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> AddToWatchlist(
        [FromBody] AddToWatchlistRequest request,
        CancellationToken cancellationToken)
    {
        var command = new AddToWatchlistCommand(
            request.ProfileId,
            request.ContentId,
            request.ContentType);

        var result = await _mediator.Send(command, cancellationToken);
        return CreatedAtAction(nameof(GetWatchlist), new { profileId = request.ProfileId }, result);
    }

    /// <summary>
    /// Remove an item from the watchlist.
    /// </summary>
    [HttpDelete("{id:guid}")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> RemoveFromWatchlist(
        Guid id,
        CancellationToken cancellationToken)
    {
        await _watchlistRepository.DeleteAsync(id, cancellationToken);
        return NoContent();
    }
}

public record AddToWatchlistRequest(
    Guid ProfileId,
    string ContentId,
    ContentType ContentType);
