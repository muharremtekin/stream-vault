namespace CatalogService.Domain.ValueObjects;

public sealed class Duration : IEquatable<Duration>
{
    public int Hours { get; private set; }
    public int Minutes { get; private set; }

    public Duration() { }

    public Duration(int hours, int minutes)
    {
        if (hours < 0)
            throw new ArgumentOutOfRangeException(nameof(hours), "Hours cannot be negative.");
        if (minutes < 0 || minutes > 59)
            throw new ArgumentOutOfRangeException(nameof(minutes), "Minutes must be between 0 and 59.");

        Hours = hours;
        Minutes = minutes;
    }

    public static Duration FromMinutes(int totalMinutes)
    {
        if (totalMinutes < 0)
            throw new ArgumentOutOfRangeException(nameof(totalMinutes), "Total minutes cannot be negative.");

        return new Duration(totalMinutes / 60, totalMinutes % 60);
    }

    public int TotalMinutes => Hours * 60 + Minutes;

    public override string ToString() => $"{Hours}h {Minutes}m";

    public bool Equals(Duration? other)
    {
        if (other is null) return false;
        return Hours == other.Hours && Minutes == other.Minutes;
    }

    public override bool Equals(object? obj) => Equals(obj as Duration);

    public override int GetHashCode() => HashCode.Combine(Hours, Minutes);

    public static bool operator ==(Duration? left, Duration? right) =>
        left is null ? right is null : left.Equals(right);

    public static bool operator !=(Duration? left, Duration? right) => !(left == right);
}
