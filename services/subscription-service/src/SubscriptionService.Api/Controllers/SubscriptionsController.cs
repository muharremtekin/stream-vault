using MediatR;
using Microsoft.AspNetCore.Mvc;
using SubscriptionService.Application.Commands.CancelSubscription;
using SubscriptionService.Application.Commands.ChangePlan;
using SubscriptionService.Application.Commands.CreateSubscription;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Queries.GetInvoices;
using SubscriptionService.Application.Queries.GetMySubscription;
using SubscriptionService.Application.Queries.GetPaymentHistory;

namespace SubscriptionService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
public class SubscriptionsController : ControllerBase
{
    private readonly IMediator _mediator;

    public SubscriptionsController(IMediator mediator)
    {
        _mediator = mediator;
    }

    [HttpPost]
    [ProducesResponseType(typeof(CreateSubscriptionResult), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> CreateSubscription(
        [FromBody] CreateSubscriptionRequest request,
        CancellationToken cancellationToken)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var command = new CreateSubscriptionCommand(userId, request.PlanId, request.CardNumber);
        var result = await _mediator.Send(command, cancellationToken);
        return StatusCode(StatusCodes.Status201Created, result);
    }

    [HttpGet("me")]
    [ProducesResponseType(typeof(SubscriptionDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetMySubscription(CancellationToken cancellationToken)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var result = await _mediator.Send(new GetMySubscriptionQuery(userId), cancellationToken);
        if (result is null)
            return NotFound(new { error = "No active subscription found." });

        return Ok(result);
    }

    [HttpPut("me/plan")]
    [ProducesResponseType(typeof(ChangePlanResult), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> ChangePlan(
        [FromBody] ChangePlanRequest request,
        CancellationToken cancellationToken)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var command = new ChangePlanCommand(userId, request.NewPlanId, request.CardNumber);
        var result = await _mediator.Send(command, cancellationToken);
        return Ok(result);
    }

    [HttpPost("me/cancel")]
    [ProducesResponseType(typeof(CancelSubscriptionResult), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> CancelSubscription(CancellationToken cancellationToken)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var result = await _mediator.Send(new CancelSubscriptionCommand(userId), cancellationToken);
        return Ok(result);
    }

    [HttpGet("me/invoices")]
    [ProducesResponseType(typeof(List<InvoiceDto>), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetInvoices(CancellationToken cancellationToken)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var invoices = await _mediator.Send(new GetInvoicesQuery(userId), cancellationToken);
        return Ok(invoices);
    }

    [HttpGet("me/payments")]
    [ProducesResponseType(typeof(List<PaymentDto>), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    public async Task<IActionResult> GetPaymentHistory(
        [FromQuery] int limit = 20,
        [FromQuery] int offset = 0,
        CancellationToken cancellationToken = default)
    {
        if (!TryGetUserId(out var userId))
            return Unauthorized(new { error = "X-User-Id header is missing or invalid." });

        var payments = await _mediator.Send(
            new GetPaymentHistoryQuery(userId, limit, offset), cancellationToken);
        return Ok(payments);
    }

    private bool TryGetUserId(out Guid userId)
    {
        userId = Guid.Empty;
        var header = Request.Headers["X-User-Id"].FirstOrDefault();
        return !string.IsNullOrEmpty(header) && Guid.TryParse(header, out userId);
    }
}

public record CreateSubscriptionRequest(Guid PlanId, string CardNumber);
public record ChangePlanRequest(Guid NewPlanId, string CardNumber);
