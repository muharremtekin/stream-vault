using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class SubscriptionConfiguration : IEntityTypeConfiguration<Subscription>
{
    public void Configure(EntityTypeBuilder<Subscription> builder)
    {
        builder.ToTable("subscriptions");

        builder.HasKey(s => s.Id)
            .HasName("pk_subscriptions");

        builder.Property(s => s.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(s => s.UserId)
            .HasColumnName("user_id")
            .IsRequired();

        builder.HasIndex(s => s.UserId)
            .HasDatabaseName("idx_subscriptions_user_id");

        builder.Property(s => s.PlanId)
            .HasColumnName("plan_id")
            .IsRequired();

        builder.Property(s => s.Status)
            .HasColumnName("status")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.HasIndex(s => s.Status)
            .HasDatabaseName("idx_subscriptions_status");

        builder.Property(s => s.PeriodStart)
            .HasColumnName("period_start")
            .IsRequired();

        builder.Property(s => s.PeriodEnd)
            .HasColumnName("period_end")
            .IsRequired();

        builder.Property(s => s.AutoRenew)
            .HasColumnName("auto_renew")
            .IsRequired();

        builder.Property(s => s.CancelledAt)
            .HasColumnName("cancelled_at");

        builder.Property(s => s.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.Property(s => s.UpdatedAt)
            .HasColumnName("updated_at")
            .IsRequired();

        builder.HasMany(s => s.Payments)
            .WithOne(p => p.Subscription)
            .HasForeignKey(p => p.SubscriptionId)
            .HasConstraintName("fk_payments_subscription_id")
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(s => s.Invoices)
            .WithOne(i => i.Subscription)
            .HasForeignKey(i => i.SubscriptionId)
            .HasConstraintName("fk_invoices_subscription_id")
            .OnDelete(DeleteBehavior.Cascade);
    }
}
