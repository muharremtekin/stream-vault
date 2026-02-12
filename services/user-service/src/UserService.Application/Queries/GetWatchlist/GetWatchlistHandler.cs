using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;

namespace UserService.Application.Queries.GetWatchlist;

public class GetWatchlistHandler : IRequestHandler<GetWatchlistQuery, List<WatchlistItemDto>>
{
    private readonly IWatchlistRepository _watchlistRepository;

    public GetWatchlistHandler(IWatchlistRepository watchlistRepository)
    {
        _watchlistRepository = watchlistRepository;
    }

    public async Task<List<WatchlistItemDto>> Handle(GetWatchlistQuery request, CancellationToken cancellationToken)
    {
        var items = await _watchlistRepository.GetByProfileIdAsync(request.ProfileId, cancellationToken);

        return items.Select(i => new WatchlistItemDto
        {
            Id = i.Id,
            ContentId = i.ContentId,
            ContentType = i.ContentType.ToString(),
            AddedAt = i.AddedAt,
            Note = i.Note
        }).ToList();
    }
}
