using System.Text.Json.Serialization;

namespace CatalogService.Infrastructure.Messaging;

public class EncodingResultMessage
{
    // Event envelope fields (Rule 3.4)
    [JsonPropertyName("event_id")]
    public string? EventId { get; set; }

    [JsonPropertyName("event_type")]
    public string? EventType { get; set; }

    [JsonPropertyName("timestamp")]
    public string? Timestamp { get; set; }

    [JsonPropertyName("source")]
    public string? Source { get; set; }

    [JsonPropertyName("correlation_id")]
    public string? CorrelationId { get; set; }

    // Data fields
    [JsonPropertyName("job_id")]
    public string JobId { get; set; } = string.Empty;

    [JsonPropertyName("content_id")]
    public string ContentId { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("outputs")]
    public List<EncodingOutputMessage> Outputs { get; set; } = new();

    [JsonPropertyName("duration_seconds")]
    public long DurationSeconds { get; set; }

    [JsonPropertyName("error_message")]
    public string? ErrorMessage { get; set; }

    [JsonPropertyName("completed_at")]
    public string CompletedAt { get; set; } = string.Empty;
}

public class EncodingOutputMessage
{
    [JsonPropertyName("quality")]
    public string Quality { get; set; } = string.Empty;

    [JsonPropertyName("width")]
    public int Width { get; set; }

    [JsonPropertyName("height")]
    public int Height { get; set; }

    [JsonPropertyName("bitrate_kbps")]
    public int BitrateKbps { get; set; }

    [JsonPropertyName("segment_count")]
    public int SegmentCount { get; set; }

    [JsonPropertyName("playlist_path")]
    public string PlaylistPath { get; set; } = string.Empty;
}
