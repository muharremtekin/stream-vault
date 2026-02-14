namespace UserService.Application.DTOs;

public record ContentRatingDto(
    Guid Id,
    string ContentId,
    int Rating,
    DateTime RatedAt
);
