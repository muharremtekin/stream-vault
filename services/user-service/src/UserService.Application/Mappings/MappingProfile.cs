using AutoMapper;
using UserService.Application.DTOs;
using UserService.Domain.Entities;

namespace UserService.Application.Mappings;

public class MappingProfile : Profile
{
    public MappingProfile()
    {
        CreateMap<User, UserDto>()
            .ForMember(dest => dest.Role, opt => opt.MapFrom(src => src.Role.ToString()))
            .ForMember(dest => dest.ProfileCount, opt => opt.MapFrom(src => src.Profiles.Count));

        CreateMap<Domain.Entities.Profile, ProfileDto>()
            .ForMember(dest => dest.Icon, opt => opt.MapFrom(src => src.Icon.ToString()));

        CreateMap<WatchlistItem, WatchlistItemDto>()
            .ForMember(dest => dest.ContentType, opt => opt.MapFrom(src => src.ContentType.ToString()));
    }
}
