namespace CatalogService.Domain.ValueObjects;

public class QualityInfo
{
    public string Label { get; private set; } = string.Empty;
    public int Width { get; private set; }
    public int Height { get; private set; }
    public int BitrateKbps { get; private set; }
    public int SegmentCount { get; private set; }

    public QualityInfo() { }

    public QualityInfo(string label, int width, int height, int bitrateKbps, int segmentCount)
    {
        Label = label ?? throw new ArgumentNullException(nameof(label));
        Width = width;
        Height = height;
        BitrateKbps = bitrateKbps;
        SegmentCount = segmentCount;
    }
}
