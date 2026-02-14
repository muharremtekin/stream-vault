using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class SagaStateConfiguration : IEntityTypeConfiguration<SagaState>
{
    public void Configure(EntityTypeBuilder<SagaState> builder)
    {
        builder.ToTable("saga_states");

        builder.HasKey(s => s.Id)
            .HasName("pk_saga_states");

        builder.Property(s => s.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(s => s.SagaType)
            .HasColumnName("saga_type")
            .HasMaxLength(100)
            .IsRequired();

        builder.Property(s => s.CurrentStep)
            .HasColumnName("current_step")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.Property(s => s.Status)
            .HasColumnName("status")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.HasIndex(s => s.Status)
            .HasDatabaseName("idx_saga_states_status");

        builder.Property(s => s.StateData)
            .HasColumnName("state_data")
            .HasColumnType("text")
            .IsRequired();

        builder.Property(s => s.ErrorMessage)
            .HasColumnName("error_message")
            .HasMaxLength(2000);

        builder.Property(s => s.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.Property(s => s.UpdatedAt)
            .HasColumnName("updated_at")
            .IsRequired();
    }
}
