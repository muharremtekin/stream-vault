namespace SubscriptionService.Api.BackgroundServices;

public class OutboxProcessorOptions
{
    public const string SectionName = "OutboxProcessor";

    /// <summary>Her döngü arasındaki bekleme süresi (saniye). Varsayılan: 5</summary>
    public int IntervalSeconds { get; set; } = 5;

    /// <summary>Bir mesajın dead-letter'a taşınmadan önce yapabileceği maksimum deneme sayısı. Varsayılan: 3</summary>
    public int MaxRetryCount { get; set; } = 3;

    /// <summary>Her döngüde işlenecek maksimum mesaj sayısı. Varsayılan: 50</summary>
    public int BatchSize { get; set; } = 50;

    /// <summary>Graceful shutdown sırasında kalan mesajları işlemek için tanınan süre (saniye). Varsayılan: 10</summary>
    public int ShutdownTimeoutSeconds { get; set; } = 10;
}
