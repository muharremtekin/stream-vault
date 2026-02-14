using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Configurations;

public class ProfileConfiguration : IEntityTypeConfiguration<Profile>
{
    public void Configure(EntityTypeBuilder<Profile> builder)
    {
        builder.ToTable("profiles");

        builder.HasKey(p => p.Id)
            .HasName("pk_profiles");

        builder.Property(p => p.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(p => p.UserId)
            .HasColumnName("user_id")
            .IsRequired();

        builder.Property(p => p.Name)
            .HasColumnName("name")
            .HasMaxLength(20)
            .IsRequired();

        builder.Property(p => p.Icon)
            .HasColumnName("icon")
            .HasConversion<string>()
            .HasMaxLength(50)
            .IsRequired();

        builder.Property(p => p.IsKids)
            .HasColumnName("is_kids")
            .IsRequired();

        builder.Property(p => p.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.HasIndex(p => p.UserId)
            .HasDatabaseName("idx_profiles_user_id");

        builder.HasMany(p => p.WatchlistItems)
            .WithOne(w => w.Profile)
            .HasForeignKey(w => w.ProfileId)
            .HasConstraintName("fk_watchlist_items_profile_id")
            .OnDelete(DeleteBehavior.Cascade);
    }
}
