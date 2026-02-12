using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Configurations;

public class WatchlistItemConfiguration : IEntityTypeConfiguration<WatchlistItem>
{
    public void Configure(EntityTypeBuilder<WatchlistItem> builder)
    {
        builder.ToTable("watchlist_items");

        builder.HasKey(w => w.Id);

        builder.Property(w => w.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(w => w.ProfileId)
            .HasColumnName("profile_id")
            .IsRequired();

        builder.Property(w => w.ContentId)
            .HasColumnName("content_id")
            .HasMaxLength(256)
            .IsRequired();

        builder.Property(w => w.ContentType)
            .HasColumnName("content_type")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.Property(w => w.AddedAt)
            .HasColumnName("added_at")
            .IsRequired();

        builder.Property(w => w.Note)
            .HasColumnName("note")
            .HasMaxLength(500);

        builder.HasIndex(w => new { w.ProfileId, w.ContentId })
            .IsUnique();
    }
}
