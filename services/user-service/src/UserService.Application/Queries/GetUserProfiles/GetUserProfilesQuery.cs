using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Queries.GetUserProfiles;

public record GetUserProfilesQuery(
    Guid UserId
) : IRequest<List<ProfileDto>>;
