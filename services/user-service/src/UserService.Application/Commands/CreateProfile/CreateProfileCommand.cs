using MediatR;
using UserService.Application.DTOs;
using UserService.Domain.Enums;

namespace UserService.Application.Commands.CreateProfile;

public record CreateProfileCommand(
    Guid UserId,
    string Name,
    ProfileIcon Icon,
    bool IsKids
) : IRequest<ProfileDto>;
