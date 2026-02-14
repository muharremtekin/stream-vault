using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;

namespace UserService.Application.Queries.GetUserRatings;

public class GetUserRatingsHandler : IRequestHandler<GetUserRatingsQuery, List<ContentRatingDto>>
{
    private readonly IContentRatingRepository _ratingRepository;

    public GetUserRatingsHandler(IContentRatingRepository ratingRepository)
    {
        _ratingRepository = ratingRepository;
    }

    public async Task<List<ContentRatingDto>> Handle(GetUserRatingsQuery request, CancellationToken cancellationToken)
    {
        var ratings = await _ratingRepository.GetByUserIdAsync(
            request.UserId, request.Page, request.PageSize, cancellationToken);

        return ratings.Select(r => new ContentRatingDto(r.Id, r.ContentId, r.Rating, r.RatedAt)).ToList();
    }
}
