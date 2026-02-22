using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;
using UserService.Domain.Exceptions;

namespace UserService.Application.Commands.CreateProfile;

public class CreateProfileHandler : IRequestHandler<CreateProfileCommand, ProfileDto>
{
    private const int MaxProfilesPerUser = 5;

    private readonly IProfileRepository _profileRepository;
    private readonly IUserRepository _userRepository;

    public CreateProfileHandler(
        IProfileRepository profileRepository,
        IUserRepository userRepository)
    {
        _profileRepository = profileRepository;
        _userRepository = userRepository;
    }

    public async Task<ProfileDto> Handle(CreateProfileCommand request, CancellationToken cancellationToken)
    {
        var user = await _userRepository.GetByIdAsync(request.UserId, cancellationToken);
        if (user is null)
        {
            throw new UserNotFoundException(request.UserId);
        }

        var profileCount = await _profileRepository.CountByUserIdAsync(request.UserId, cancellationToken);
        if (profileCount >= MaxProfilesPerUser)
        {
            throw new InvalidOperationException(
                $"Maximum number of profiles ({MaxProfilesPerUser}) has been reached for this user.");
        }

        var profile = new Profile
        {
            Id = Guid.NewGuid(),
            UserId = request.UserId,
            Name = request.Name,
            Icon = request.Icon,
            IsKids = request.IsKids,
            CreatedAt = DateTime.UtcNow
        };

        await _profileRepository.AddAsync(profile, cancellationToken);

        return new ProfileDto
        {
            Id = profile.Id,
            Name = profile.Name,
            Icon = profile.Icon.ToString().ToLowerInvariant(),
            IsKids = profile.IsKids
        };
    }
}
