namespace CatalogService.Domain.ValueObjects;

public sealed class ThumbnailUrl : IEquatable<ThumbnailUrl>
{
    public string Value { get; }

    public ThumbnailUrl() => Value = string.Empty;

    public ThumbnailUrl(string url)
    {
        if (string.IsNullOrWhiteSpace(url))
            throw new ArgumentException("Thumbnail URL cannot be empty.", nameof(url));

        if (!Uri.TryCreate(url, UriKind.Absolute, out var uri))
            throw new ArgumentException("Thumbnail URL must be a valid absolute URI.", nameof(url));

        if (uri.Scheme != "http" && uri.Scheme != "https")
            throw new ArgumentException("Thumbnail URL must use HTTP or HTTPS scheme.", nameof(url));

        Value = url;
    }

    public static implicit operator string(ThumbnailUrl thumbnailUrl) => thumbnailUrl.Value;
    public static explicit operator ThumbnailUrl(string url) => new(url);

    public override string ToString() => Value;

    public bool Equals(ThumbnailUrl? other)
    {
        if (other is null) return false;
        return string.Equals(Value, other.Value, StringComparison.OrdinalIgnoreCase);
    }

    public override bool Equals(object? obj) => Equals(obj as ThumbnailUrl);

    public override int GetHashCode() => Value.ToLowerInvariant().GetHashCode();

    public static bool operator ==(ThumbnailUrl? left, ThumbnailUrl? right) =>
        left is null ? right is null : left.Equals(right);

    public static bool operator !=(ThumbnailUrl? left, ThumbnailUrl? right) => !(left == right);
}
