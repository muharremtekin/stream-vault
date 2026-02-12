using AutoMapper;
using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using CatalogService.Domain.Enums;
using MediatR;

namespace CatalogService.Application.Queries.GetContentById;

public class GetContentByIdHandler : IRequestHandler<GetContentByIdQuery, ContentSummaryDto?>
{
    private readonly IMovieRepository _movieRepository;
    private readonly ISeriesRepository _seriesRepository;
    private readonly IMapper _mapper;

    public GetContentByIdHandler(
        IMovieRepository movieRepository,
        ISeriesRepository seriesRepository,
        IMapper mapper)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
        _seriesRepository = seriesRepository ?? throw new ArgumentNullException(nameof(seriesRepository));
        _mapper = mapper ?? throw new ArgumentNullException(nameof(mapper));
    }

    public async Task<ContentSummaryDto?> Handle(GetContentByIdQuery request, CancellationToken cancellationToken)
    {
        switch (request.ContentType)
        {
            case ContentType.Movie:
                var movie = await _movieRepository.GetByIdAsync(request.Id, cancellationToken);
                return movie is null ? null : _mapper.Map<ContentSummaryDto>(movie);

            case ContentType.Series:
                var series = await _seriesRepository.GetByIdAsync(request.Id, cancellationToken);
                return series is null ? null : _mapper.Map<ContentSummaryDto>(series);

            default:
                throw new ArgumentOutOfRangeException(nameof(request.ContentType), "Invalid content type.");
        }
    }
}
