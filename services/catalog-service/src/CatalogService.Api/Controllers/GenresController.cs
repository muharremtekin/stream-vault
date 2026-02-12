using AutoMapper;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using CatalogService.Application.Queries.GetByGenre;
using CatalogService.Application.Queries.GetMovies;
using MediatR;
using Microsoft.AspNetCore.Mvc;

namespace CatalogService.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
public class GenresController : ControllerBase
{
    private readonly IMediator _mediator;
    private readonly IGenreRepository _genreRepository;
    private readonly IMovieRepository _movieRepository;
    private readonly ISeriesRepository _seriesRepository;
    private readonly IMapper _mapper;

    public GenresController(
        IMediator mediator,
        IGenreRepository genreRepository,
        IMovieRepository movieRepository,
        ISeriesRepository seriesRepository,
        IMapper mapper)
    {
        _mediator = mediator ?? throw new ArgumentNullException(nameof(mediator));
        _genreRepository = genreRepository ?? throw new ArgumentNullException(nameof(genreRepository));
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    /// <summary>
    /// Gets all genres with content counts.
    /// </summary>
    [HttpGet]
    [ProducesResponseType(typeof(List<GenreDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetAllGenres(CancellationToken cancellationToken = default)
    {
        var genres = await _genreRepository.GetAllAsync(cancellationToken);
        var genreDtos = new List<GenreDto>();

        foreach (var genre in genres)
        {
            var movieCount = await _movieRepository.CountAsync(genre: genre.Name, cancellationToken: cancellationToken);
            var seriesCount = await _seriesRepository.CountAsync(genre: genre.Name, cancellationToken: cancellationToken);

            var dto = _mapper.Map<GenreDto>(genre) with
            {
                ContentCount = movieCount + seriesCount
            };

            genreDtos.Add(dto);
        }

        return Ok(genreDtos);
    }

    /// <summary>
    /// Gets a genre by slug with paginated content.
    /// </summary>
    [HttpGet("{slug}")]
    [ProducesResponseType(typeof(object), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetBySlug(
        string slug,
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 20,
        CancellationToken cancellationToken = default)
    {
        var genre = await _genreRepository.GetBySlugAsync(slug, cancellationToken);

        if (genre is null)
            return NotFound(new { message = $"Genre with slug '{slug}' not found." });

        var query = new GetByGenreQuery
        {
            Slug = slug,
            Page = page,
            PageSize = pageSize
        };

        var content = await _mediator.Send(query, cancellationToken);

        var genreDto = _mapper.Map<GenreDto>(genre) with
        {
            ContentCount = content.TotalCount
        };

        return Ok(new
        {
            Genre = genreDto,
            Content = content
        });
    }
}
