using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;
using UserService.Domain.Entities;

namespace UserService.Infrastructure.Persistence.Configurations;

public class ContentRatingConfiguration : IEntityTypeConfiguration<ContentRating>
{
    public void Configure(EntityTypeBuilder<ContentRating> builder)
    {
        builder.ToTable("content_ratings");

        builder.HasKey(r => r.Id)
            .HasName("pk_content_ratings");

        builder.Property(r => r.Id)
            .HasColumnName("id")
            .ValueGeneratedNever();

        builder.Property(r => r.UserId)
            .HasColumnName("user_id")
            .IsRequired();

        builder.Property(r => r.ContentId)
            .HasColumnName("content_id")
            .HasMaxLength(256)
            .IsRequired();

        builder.Property(r => r.Rating)
            .HasColumnName("rating")
            .IsRequired();

        builder.Property(r => r.RatedAt)
            .HasColumnName("rated_at")
            .IsRequired();

        builder.HasOne(r => r.User)
            .WithMany()
            .HasForeignKey(r => r.UserId)
            .HasConstraintName("fk_content_ratings_user_id")
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasIndex(r => new { r.UserId, r.ContentId })
            .IsUnique()
            .HasDatabaseName("idx_content_ratings_user_id_content_id");

        builder.HasIndex(r => r.ContentId)
            .HasDatabaseName("idx_content_ratings_content_id");
    }
}
