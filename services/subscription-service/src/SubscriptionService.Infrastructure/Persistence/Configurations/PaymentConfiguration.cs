using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class PaymentConfiguration : IEntityTypeConfiguration<Payment>
{
    public void Configure(EntityTypeBuilder<Payment> builder)
    {
        builder.ToTable("payments");

        builder.HasKey(p => p.Id)
            .HasName("pk_payments");

        builder.Property(p => p.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(p => p.SubscriptionId)
            .HasColumnName("subscription_id")
            .IsRequired();

        builder.HasIndex(p => p.SubscriptionId)
            .HasDatabaseName("idx_payments_subscription_id");

        builder.HasIndex(p => new { p.SubscriptionId, p.CreatedAt })
            .HasDatabaseName("idx_payments_subscription_created_at");

        builder.Property(p => p.Amount)
            .HasColumnName("amount")
            .HasColumnType("decimal(10,2)")
            .IsRequired();

        builder.Property(p => p.Currency)
            .HasColumnName("currency")
            .HasMaxLength(10)
            .IsRequired();

        builder.Property(p => p.Status)
            .HasColumnName("status")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.Property(p => p.TransactionId)
            .HasColumnName("transaction_id")
            .HasMaxLength(256);

        builder.HasIndex(p => p.TransactionId)
            .HasDatabaseName("idx_payments_transaction_id");

        builder.Property(p => p.FailureReason)
            .HasColumnName("failure_reason")
            .HasMaxLength(500);

        builder.Property(p => p.CardLastFour)
            .HasColumnName("card_last_four")
            .HasMaxLength(4)
            .IsRequired();

        builder.Property(p => p.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();
    }
}
