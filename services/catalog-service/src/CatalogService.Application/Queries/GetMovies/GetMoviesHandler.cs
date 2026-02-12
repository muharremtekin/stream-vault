using AutoMapper;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using MediatR;

namespace CatalogService.Application.Queries.GetMovies;

public class GetMoviesHandler : IRequestHandler<GetMoviesQuery, PagedResult<MovieDto>>
{
    private readonly IMovieRepository _movieRepository;
    private readonly IMapper _mapper;

    public GetMoviesHandler(IMovieRepository movieRepository, IMapper mapper)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    public async Task<PagedResult<MovieDto>> Handle(GetMoviesQuery request, CancellationToken cancellationToken)
    {
        var movies = await _movieRepository.GetAllAsync(
            page: request.Page,
            pageSize: request.PageSize,
            genre: request.Genre,
            sort: request.Sort,
            year: request.Year,
            cancellationToken: cancellationToken
        );

        var totalCount = await _movieRepository.CountAsync(
            genre: request.Genre,
            year: request.Year,
            cancellationToken: cancellationToken
        );

        var movieDtos = _mapper.Map<List<MovieDto>>(movies);

        return new PagedResult<MovieDto>
        {
            Items = movieDtos,
            TotalCount = totalCount,
            Page = request.Page,
            PageSize = request.PageSize
        };
    }
}
