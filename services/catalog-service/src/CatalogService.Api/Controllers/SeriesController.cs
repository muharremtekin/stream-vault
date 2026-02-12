using AutoMapper;
using CatalogService.Application.Commands.AddEpisode;
using CatalogService.Application.Commands.CreateSeries;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using CatalogService.Application.Queries.GetContentById;
using CatalogService.Domain.Enums;
using MediatR;
using Microsoft.AspNetCore.Mvc;

namespace CatalogService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
public class SeriesController : ControllerBase
{
    private readonly IMediator _mediator;
    private readonly ISeriesRepository _seriesRepository;
    private readonly IMapper _mapper;

    public SeriesController(IMediator mediator, ISeriesRepository seriesRepository, IMapper mapper)
    {
        _mediator = mediator ?? throw new ArgumentNullException(nameof(mediator));
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    /// <summary>
    /// Gets a paginated list of series with optional filtering.
    /// </summary>
    [HttpGet]
    [ProducesResponseType(typeof(object), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetSeries(
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 20,
        [FromQuery] string? genre = null,
        [FromQuery] string? sort = null,
        [FromQuery] int? year = null,
        CancellationToken cancellationToken = default)
    {
        var seriesList = await _seriesRepository.GetAllAsync(page, pageSize, genre, sort, year, cancellationToken);
        var totalCount = await _seriesRepository.CountAsync(genre, year, cancellationToken);
        var seriesDtos = _mapper.Map<List<SeriesDto>>(seriesList);

        var result = new
        {
            Items = seriesDtos,
            TotalCount = totalCount,
            Page = page,
            PageSize = pageSize,
            TotalPages = (int)Math.Ceiling(totalCount / (double)pageSize),
            HasNextPage = page < (int)Math.Ceiling(totalCount / (double)pageSize),
            HasPreviousPage = page > 1
        };

        return Ok(result);
    }

    /// <summary>
    /// Gets a series by its ID.
    /// </summary>
    [HttpGet("{id}")]
    [ProducesResponseType(typeof(SeriesDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetSeriesById(
        string id,
        CancellationToken cancellationToken = default)
    {
        var series = await _seriesRepository.GetByIdAsync(id, cancellationToken);

        if (series is null)
            return NotFound(new { message = $"Series with ID '{id}' not found." });

        var seriesDto = _mapper.Map<SeriesDto>(series);
        return Ok(seriesDto);
    }

    /// <summary>
    /// Creates a new series (admin only).
    /// </summary>
    [HttpPost]
    [ProducesResponseType(typeof(object), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> CreateSeries(
        [FromBody] CreateSeriesCommand command,
        CancellationToken cancellationToken = default)
    {
        var seriesId = await _mediator.Send(command, cancellationToken);

        return CreatedAtAction(
            nameof(GetSeriesById),
            new { id = seriesId },
            new { id = seriesId });
    }

    /// <summary>
    /// Adds an episode to a series season.
    /// </summary>
    [HttpPost("{seriesId}/seasons/{seasonNumber}/episodes")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> AddEpisode(
        string seriesId,
        int seasonNumber,
        [FromBody] AddEpisodeRequest request,
        CancellationToken cancellationToken = default)
    {
        var command = new AddEpisodeCommand
        {
            SeriesId = seriesId,
            SeasonNumber = seasonNumber,
            EpisodeNumber = request.EpisodeNumber,
            Title = request.Title,
            Description = request.Description,
            DurationMinutes = request.DurationMinutes,
            ThumbnailUrl = request.ThumbnailUrl
        };

        await _mediator.Send(command, cancellationToken);

        return Ok(new { message = "Episode added successfully." });
    }
}

public record AddEpisodeRequest
{
    public int EpisodeNumber { get; init; }
    public string Title { get; init; } = string.Empty;
    public string Description { get; init; } = string.Empty;
    public int DurationMinutes { get; init; }
    public string ThumbnailUrl { get; init; } = string.Empty;
}
