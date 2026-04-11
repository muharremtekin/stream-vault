using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class OutboxMessageConfiguration : IEntityTypeConfiguration<OutboxMessage>
{
    public void Configure(EntityTypeBuilder<OutboxMessage> builder)
    {
        builder.ToTable("outbox_messages");

        builder.HasKey(o => o.Id)
            .HasName("pk_outbox_messages");

        builder.Property(o => o.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(o => o.EventType)
            .HasColumnName("event_type")
            .HasMaxLength(200)
            .IsRequired();

        builder.Property(o => o.Payload)
            .HasColumnName("payload")
            .HasColumnType("text")
            .IsRequired();

        builder.Property(o => o.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.Property(o => o.NextAttemptAt)
            .HasColumnName("next_attempt_at");

        builder.Property(o => o.ProcessedAt)
            .HasColumnName("processed_at");

        builder.HasIndex(o => o.ProcessedAt)
            .HasDatabaseName("idx_outbox_messages_processed_at");

        builder.Property(o => o.RetryCount)
            .HasColumnName("retry_count")
            .IsRequired();

        builder.Property(o => o.ErrorMessage)
            .HasColumnName("error_message")
            .HasMaxLength(2000);

        builder.Property(o => o.LastAttemptedAt)
            .HasColumnName("last_attempted_at");

        builder.Property(o => o.IsDeadLetter)
            .HasColumnName("is_dead_letter")
            .HasDefaultValue(false)
            .IsRequired();

        builder.HasIndex(o => o.IsDeadLetter)
            .HasDatabaseName("idx_outbox_messages_is_dead_letter");

        builder.HasIndex(o => new { o.NextAttemptAt, o.CreatedAt })
            .HasDatabaseName("idx_outbox_messages_due_polling")
            .HasFilter("\"processed_at\" IS NULL AND \"is_dead_letter\" = false");
    }
}
