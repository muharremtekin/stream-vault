using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Infrastructure.Persistence.Configurations;

public class InvoiceConfiguration : IEntityTypeConfiguration<Invoice>
{
    public void Configure(EntityTypeBuilder<Invoice> builder)
    {
        builder.ToTable("invoices");

        builder.HasKey(i => i.Id)
            .HasName("pk_invoices");

        builder.Property(i => i.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(i => i.SubscriptionId)
            .HasColumnName("subscription_id")
            .IsRequired();

        builder.HasIndex(i => i.SubscriptionId)
            .HasDatabaseName("idx_invoices_subscription_id");

        builder.Property(i => i.InvoiceNumber)
            .HasColumnName("invoice_number")
            .HasMaxLength(50)
            .IsRequired();

        builder.HasIndex(i => i.InvoiceNumber)
            .IsUnique()
            .HasDatabaseName("idx_invoices_invoice_number");

        builder.Property(i => i.Amount)
            .HasColumnName("amount")
            .HasColumnType("decimal(10,2)")
            .IsRequired();

        builder.Property(i => i.Currency)
            .HasColumnName("currency")
            .HasMaxLength(10)
            .IsRequired();

        builder.Property(i => i.PeriodStart)
            .HasColumnName("period_start")
            .IsRequired();

        builder.Property(i => i.PeriodEnd)
            .HasColumnName("period_end")
            .IsRequired();

        builder.Property(i => i.IssuedAt)
            .HasColumnName("issued_at")
            .IsRequired();
    }
}
