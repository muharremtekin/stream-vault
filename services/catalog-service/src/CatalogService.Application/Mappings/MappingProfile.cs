using AutoMapper;
using CatalogService.Application.DTOs;
using CatalogService.Domain.Entities;
using CatalogService.Domain.Enums;

namespace CatalogService.Application.Mappings;

public class MappingProfile : Profile
{
    public MappingProfile()
    {
        // Movie -> MovieDto
        CreateMap<Movie, MovieDto>()
            .ForMember(dest => dest.DurationMinutes, opt => opt.MapFrom(src => src.Duration.TotalMinutes))
            .ForMember(dest => dest.DurationFormatted, opt => opt.MapFrom(src => src.Duration.ToString()))
            .ForMember(dest => dest.MaturityRating, opt => opt.MapFrom(src => src.MaturityRating.ToString()))
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()))
            .ForMember(dest => dest.VideoStatus, opt => opt.MapFrom(src => src.VideoStatus.ToString()));

        // Movie -> ContentSummaryDto
        CreateMap<Movie, ContentSummaryDto>()
            .ForMember(dest => dest.ContentType, opt => opt.MapFrom(_ => ContentType.Movie))
            .ForMember(dest => dest.MaturityRating, opt => opt.MapFrom(src => src.MaturityRating.ToString()))
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()));

        // Series -> SeriesDto
        CreateMap<Series, SeriesDto>()
            .ForMember(dest => dest.MaturityRating, opt => opt.MapFrom(src => src.MaturityRating.ToString()))
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()))
            .ForMember(dest => dest.TotalSeasons, opt => opt.MapFrom(src => src.Seasons.Count))
            .ForMember(dest => dest.TotalEpisodes, opt => opt.MapFrom(src =>
                src.Seasons.Sum(s => s.Episodes.Count)));

        // Series -> ContentSummaryDto
        CreateMap<Series, ContentSummaryDto>()
            .ForMember(dest => dest.ContentType, opt => opt.MapFrom(_ => ContentType.Series))
            .ForMember(dest => dest.MaturityRating, opt => opt.MapFrom(src => src.MaturityRating.ToString()))
            .ForMember(dest => dest.AverageRating, opt => opt.Ignore())
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()));

        // Season -> SeasonDto
        CreateMap<Season, SeasonDto>()
            .ForMember(dest => dest.EpisodeCount, opt => opt.MapFrom(src => src.Episodes.Count));

        // Episode -> EpisodeDto
        CreateMap<Episode, EpisodeDto>()
            .ForMember(dest => dest.DurationMinutes, opt => opt.MapFrom(src => src.Duration.TotalMinutes))
            .ForMember(dest => dest.DurationFormatted, opt => opt.MapFrom(src => src.Duration.ToString()))
            .ForMember(dest => dest.VideoStatus, opt => opt.MapFrom(src => src.VideoStatus.ToString()));

        // CastMember -> CastMemberDto
        CreateMap<CastMember, CastMemberDto>();

        // Genre -> GenreDto (ContentCount set manually)
        CreateMap<Genre, GenreDto>()
            .ForMember(dest => dest.ContentCount, opt => opt.Ignore());
    }
}
