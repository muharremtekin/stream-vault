using SubscriptionService.Domain.Enums;

namespace SubscriptionService.Domain.Entities;

public class SagaState
{
    public Guid Id { get; set; } = Guid.NewGuid();
    public string SagaType { get; set; } = string.Empty;
    public SagaStep CurrentStep { get; set; }
    public SagaStatus Status { get; set; } = SagaStatus.Started;
    public string StateData { get; set; } = "{}";
    public string? ErrorMessage { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
