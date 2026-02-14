using MediatR;
using UserService.Application.DTOs;

namespace UserService.Application.Queries.GetUserRatings;

public record GetUserRatingsQuery(Guid UserId, int Page = 1, int PageSize = 20) : IRequest<List<ContentRatingDto>>;
