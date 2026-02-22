using AutoMapper;
using CatalogService.Application.Commands.CreateMovie;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using CatalogService.Application.Queries.GetMovies;
using CatalogService.Application.Queries.GetStreamingInfo;
using MediatR;
using Microsoft.AspNetCore.Mvc;

namespace CatalogService.Api.Controllers;

[ApiController]
[Route("api/catalog/[controller]")]
public class MoviesController : ControllerBase
{
    private readonly IMediator _mediator;
    private readonly IMovieRepository _movieRepository;
    private readonly IMapper _mapper;

    public MoviesController(IMediator mediator, IMovieRepository movieRepository, IMapper mapper)
    {
        _mediator = mediator ?? throw new ArgumentNullException(nameof(mediator));
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    /// <summary>
    /// Gets a paginated list of movies with optional filtering and sorting.
    /// </summary>
    [HttpGet]
    [ProducesResponseType(typeof(PagedResult<MovieDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetMovies(
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 20,
        [FromQuery] string? genre = null,
        [FromQuery] string? sort = null,
        [FromQuery] int? year = null,
        CancellationToken cancellationToken = default)
    {
        var query = new GetMoviesQuery
        {
            Page = page,
            PageSize = pageSize,
            Genre = genre,
            Sort = sort,
            Year = year
        };

        var result = await _mediator.Send(query, cancellationToken);
        return Ok(result);
    }

    /// <summary>
    /// Gets a movie by its ID.
    /// </summary>
    [HttpGet("{id}")]
    [ProducesResponseType(typeof(MovieDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetMovieById(
        string id,
        CancellationToken cancellationToken = default)
    {
        var movie = await _movieRepository.GetByIdAsync(id, cancellationToken);

        if (movie is null)
            return NotFound(new { message = $"Movie with ID '{id}' not found." });

        var movieDto = _mapper.Map<MovieDto>(movie);
        return Ok(movieDto);
    }

    /// <summary>
    /// Creates a new movie (admin only).
    /// </summary>
    [HttpPost]
    [ProducesResponseType(typeof(object), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    public async Task<IActionResult> CreateMovie(
        [FromBody] CreateMovieCommand command,
        CancellationToken cancellationToken = default)
    {
        var movieId = await _mediator.Send(command, cancellationToken);

        return CreatedAtAction(
            nameof(GetMovieById),
            new { id = movieId },
            new { id = movieId });
    }

    /// <summary>
    /// Gets streaming info for a movie.
    /// </summary>
    [HttpGet("{id}/streaming-info")]
    [ProducesResponseType(typeof(StreamingInfoDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetStreamingInfo(
        string id,
        CancellationToken cancellationToken = default)
    {
        var query = new GetStreamingInfoQuery { MovieId = id };
        var result = await _mediator.Send(query, cancellationToken);

        if (result is null)
            return NotFound(new { message = $"Movie with ID '{id}' not found." });

        return Ok(result);
    }
}
