using MediatR;
using UserService.Application.DTOs;
using UserService.Application.Interfaces;
using UserService.Domain.Entities;

namespace UserService.Application.Commands.AddToWatchlist;

public class AddToWatchlistHandler : IRequestHandler<AddToWatchlistCommand, WatchlistItemDto>
{
    private readonly IWatchlistRepository _watchlistRepository;
    private readonly IProfileRepository _profileRepository;

    public AddToWatchlistHandler(
        IWatchlistRepository watchlistRepository,
        IProfileRepository profileRepository)
    {
        _watchlistRepository = watchlistRepository;
        _profileRepository = profileRepository;
    }

    public async Task<WatchlistItemDto> Handle(AddToWatchlistCommand request, CancellationToken cancellationToken)
    {
        var profileExists = await _profileRepository.ExistsAsync(request.ProfileId, cancellationToken);
        if (!profileExists)
        {
            throw new InvalidOperationException($"Profile with ID '{request.ProfileId}' was not found.");
        }

        var watchlistItem = new WatchlistItem
        {
            Id = Guid.NewGuid(),
            ProfileId = request.ProfileId,
            ContentId = request.ContentId,
            ContentType = request.ContentType,
            AddedAt = DateTime.UtcNow
        };

        await _watchlistRepository.AddAsync(watchlistItem, cancellationToken);

        return new WatchlistItemDto
        {
            Id = watchlistItem.Id,
            ContentId = watchlistItem.ContentId,
            ContentType = watchlistItem.ContentType.ToString(),
            AddedAt = watchlistItem.AddedAt,
            Note = watchlistItem.Note
        };
    }
}
