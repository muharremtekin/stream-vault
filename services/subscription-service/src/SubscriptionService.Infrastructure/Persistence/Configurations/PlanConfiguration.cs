using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class PlanConfiguration : IEntityTypeConfiguration<Plan>
{
    public void Configure(EntityTypeBuilder<Plan> builder)
    {
        builder.ToTable("plans");

        builder.HasKey(p => p.Id)
            .HasName("pk_plans");

        builder.Property(p => p.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(p => p.Name)
            .HasColumnName("name")
            .HasMaxLength(100)
            .IsRequired();

        builder.HasIndex(p => p.Name)
            .IsUnique()
            .HasDatabaseName("idx_plans_name");

        builder.Property(p => p.Tier)
            .HasColumnName("tier")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.HasIndex(p => p.Tier)
            .IsUnique()
            .HasDatabaseName("idx_plans_tier");

        builder.Property(p => p.PriceMonthly)
            .HasColumnName("price_monthly")
            .HasColumnType("decimal(10,2)")
            .IsRequired();

        builder.Property(p => p.MaxScreens)
            .HasColumnName("max_screens")
            .IsRequired();

        builder.Property(p => p.MaxQuality)
            .HasColumnName("max_quality")
            .HasMaxLength(20)
            .IsRequired();

        builder.Property(p => p.Features)
            .HasColumnName("features")
            .HasMaxLength(2000)
            .IsRequired();

        builder.Property(p => p.IsActive)
            .HasColumnName("is_active")
            .IsRequired();

        builder.Property(p => p.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.Property(p => p.UpdatedAt)
            .HasColumnName("updated_at")
            .IsRequired();

        builder.HasMany(p => p.Subscriptions)
            .WithOne(s => s.Plan)
            .HasForeignKey(s => s.PlanId)
            .HasConstraintName("fk_subscriptions_plan_id")
            .OnDelete(DeleteBehavior.Restrict);
    }
}
