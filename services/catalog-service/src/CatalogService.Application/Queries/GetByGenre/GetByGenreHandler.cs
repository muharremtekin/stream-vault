using AutoMapper;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using CatalogService.Application.Queries.GetMovies;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Queries.GetByGenre;

public class GetByGenreHandler : IRequestHandler<GetByGenreQuery, PagedResult<ContentSummaryDto>>
{
    private readonly IMovieRepository _movieRepository;
    private readonly ISeriesRepository _seriesRepository;
    private readonly IGenreRepository _genreRepository;
    private readonly IMapper _mapper;

    public GetByGenreHandler(
        IMovieRepository movieRepository,
        ISeriesRepository seriesRepository,
        IGenreRepository genreRepository,
        IMapper mapper)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _genreRepository = genreRepository ?? throw new ArgumentNullException(nameof(genreRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    public async Task<PagedResult<ContentSummaryDto>> Handle(GetByGenreQuery request, CancellationToken cancellationToken)
    {
        var genre = await _genreRepository.GetBySlugAsync(request.Slug, cancellationToken);

        if (genre is null)
            throw new KeyNotFoundException($"Genre with slug '{request.Slug}' was not found.");

        var movies = await _movieRepository.GetAllAsync(
            page: 1,
            pageSize: int.MaxValue,
            genre: genre.Name,
            cancellationToken: cancellationToken
        );

        var series = await _seriesRepository.GetAllAsync(
            page: 1,
            pageSize: int.MaxValue,
            genre: genre.Name,
            cancellationToken: cancellationToken
        );

        var allContent = new List<ContentSummaryDto>();
        allContent.AddRange(_mapper.Map<List<ContentSummaryDto>>(movies));
        allContent.AddRange(_mapper.Map<List<ContentSummaryDto>>(series));

        var totalCount = allContent.Count;

        var pagedContent = allContent
            .OrderByDescending(c => c.ReleaseYear)
            .Skip((request.Page - 1) * request.PageSize)
            .Take(request.PageSize)
            .ToList();

        return new PagedResult<ContentSummaryDto>
        {
            Items = pagedContent,
            TotalCount = totalCount,
            Page = request.Page,
            PageSize = request.PageSize
        };
    }
}
