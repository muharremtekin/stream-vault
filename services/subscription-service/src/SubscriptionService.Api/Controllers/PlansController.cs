using MediatR;
using Microsoft.AspNetCore.Mvc;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Queries.GetPlans;

namespace SubscriptionService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
public class PlansController : ControllerBase
{
    private readonly IMediator _mediator;

    public PlansController(IMediator mediator)
    {
        _mediator = mediator;
    }

    [HttpGet]
    [ProducesResponseType(typeof(List<PlanDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetPlans(CancellationToken cancellationToken)
    {
        var plans = await _mediator.Send(new GetPlansQuery(), cancellationToken);
        return Ok(plans);
    }
}
