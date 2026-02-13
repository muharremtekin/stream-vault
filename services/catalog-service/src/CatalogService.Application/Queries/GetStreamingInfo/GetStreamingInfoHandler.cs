using CatalogService.Application.DTOs;
using CatalogService.Application.Interfaces;
using MediatR;

namespace CatalogService.Application.Queries.GetStreamingInfo;

public class GetStreamingInfoHandler : IRequestHandler<GetStreamingInfoQuery, StreamingInfoDto?>
{
    private readonly IMovieRepository _movieRepository;

    public GetStreamingInfoHandler(IMovieRepository movieRepository)
    {
        _movieRepository = movieRepository ?? throw new ArgumentNullException(nameof(movieRepository));
    }

    public async Task<StreamingInfoDto?> Handle(GetStreamingInfoQuery request, CancellationToken cancellationToken)
    {
        var movie = await _movieRepository.GetByIdAsync(request.MovieId, cancellationToken);

        if (movie is null)
            return null;

        var dto = new StreamingInfoDto
        {
            VideoStatus = movie.VideoStatus.ToString()
        };

        if (movie.StreamingInfo is not null)
        {
            dto = dto with
            {
                DurationSeconds = movie.StreamingInfo.DurationSeconds,
                AvailableQualities = movie.StreamingInfo.AvailableQualities
                    .Select(q => new QualityInfoDto
                    {
                        Label = q.Label,
                        Width = q.Width,
                        Height = q.Height,
                        BitrateKbps = q.BitrateKbps,
                        SegmentCount = q.SegmentCount
                    }).ToList(),
                ManifestUrl = movie.StreamingInfo.ManifestPath,
                ThumbnailUrl = movie.StreamingInfo.ThumbnailPath,
                PosterUrl = movie.StreamingInfo.PosterPath,
                EncodedAt = movie.StreamingInfo.EncodedAt
            };
        }

        return dto;
    }
}
