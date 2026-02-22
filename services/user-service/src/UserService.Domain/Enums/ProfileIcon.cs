using System.Text.Json.Serialization;

namespace UserService.Domain.Enums;

[JsonConverter(typeof(JsonStringEnumConverter<ProfileIcon>))]
public enum ProfileIcon
{
    Smile = 1,
    Cat = 2,
    Dog = 3,
    Bird = 4,
    Fish = 5,
    Rabbit = 6,
    Star = 7,
    Heart = 8,
    Ghost = 9,
    Rocket = 10
}
