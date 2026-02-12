using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Exceptions;

namespace UserService.Application.Queries.GetUserProfiles;

public class GetUserProfilesHandler : IRequestHandler<GetUserProfilesQuery, List<ProfileDto>>
{
    private readonly IProfileRepository _profileRepository;
    private readonly IUserRepository _userRepository;

    public GetUserProfilesHandler(
        IProfileRepository profileRepository,
        IUserRepository userRepository)
    {
        _profileRepository = profileRepository;
        _userRepository = userRepository;
    }

    public async Task<List<ProfileDto>> Handle(GetUserProfilesQuery request, CancellationToken cancellationToken)
    {
        var user = await _userRepository.GetByIdAsync(request.UserId, cancellationToken);
        if (user is null)
        {
            throw new UserNotFoundException(request.UserId);
        }

        var profiles = await _profileRepository.GetByUserIdAsync(request.UserId, cancellationToken);

        return profiles.Select(p => new ProfileDto
        {
            Id = p.Id,
            Name = p.Name,
            Icon = p.Icon.ToString(),
            IsKids = p.IsKids
        }).ToList();
    }
}
