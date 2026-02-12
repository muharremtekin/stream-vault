namespace CatalogService.Domain.Entities;

public class CastMember
{
    public string Name { get; set; } = string.Empty;

    public string Role { get; set; } = string.Empty;

    public string? PhotoUrl { get; set; }

    public CastMember() { }

    public CastMember(string name, string role, string? photoUrl = null)
    {
        Name = name ?? throw new ArgumentNullException(nameof(name));
        Role = role ?? throw new ArgumentNullException(nameof(role));
        PhotoUrl = photoUrl;
    }
}
