using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Commands.RateContent;

public record RateContentCommand(Guid UserId, string ContentId, int Rating) : IRequest<ContentRatingDto>;
