using System.Text.Json.Serialization;

namespace CatalogService.Domain.Enums;

[JsonConverter(typeof(JsonStringEnumConverter))]
public enum MaturityRating
{
    G = 0,
    PG = 1,
    PG13 = 2,
    R = 3,
    NC17 = 4
}
