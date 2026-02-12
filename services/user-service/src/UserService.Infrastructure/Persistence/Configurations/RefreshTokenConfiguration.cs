using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Configurations;

public class RefreshTokenConfiguration : IEntityTypeConfiguration<RefreshToken>
{
    public void Configure(EntityTypeBuilder<RefreshToken> builder)
    {
        builder.ToTable("refresh_tokens");

        builder.HasKey(rt => rt.Id);

        builder.Property(rt => rt.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(rt => rt.UserId)
            .HasColumnName("user_id")
            .IsRequired();

        builder.HasIndex(rt => rt.UserId);

        builder.Property(rt => rt.Token)
            .HasColumnName("token")
            .HasMaxLength(500)
            .IsRequired();

        builder.HasIndex(rt => rt.Token)
            .IsUnique();

        builder.Property(rt => rt.ExpiresAt)
            .HasColumnName("expires_at")
            .IsRequired();

        builder.Property(rt => rt.CreatedAt)
            .HasColumnName("created_at")
            .IsRequired();

        builder.Property(rt => rt.IsRevoked)
            .HasColumnName("is_revoked")
            .IsRequired();
    }
}
