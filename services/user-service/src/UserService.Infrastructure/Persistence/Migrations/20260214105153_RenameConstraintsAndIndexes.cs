using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace UserService.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class RenameConstraintsAndIndexes : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            // PK renames (RENAME CONSTRAINT is non-destructive)
            migrationBuilder.Sql("ALTER TABLE users RENAME CONSTRAINT \"PK_users\" TO \"pk_users\";");
            migrationBuilder.Sql("ALTER TABLE profiles RENAME CONSTRAINT \"PK_profiles\" TO \"pk_profiles\";");
            migrationBuilder.Sql("ALTER TABLE refresh_tokens RENAME CONSTRAINT \"PK_refresh_tokens\" TO \"pk_refresh_tokens\";");
            migrationBuilder.Sql("ALTER TABLE watchlist_items RENAME CONSTRAINT \"PK_watchlist_items\" TO \"pk_watchlist_items\";");

            // FK renames
            migrationBuilder.Sql("ALTER TABLE profiles RENAME CONSTRAINT \"FK_profiles_users_user_id\" TO \"fk_profiles_user_id\";");
            migrationBuilder.Sql("ALTER TABLE refresh_tokens RENAME CONSTRAINT \"FK_refresh_tokens_users_user_id\" TO \"fk_refresh_tokens_user_id\";");
            migrationBuilder.Sql("ALTER TABLE watchlist_items RENAME CONSTRAINT \"FK_watchlist_items_profiles_profile_id\" TO \"fk_watchlist_items_profile_id\";");

            // Index renames
            migrationBuilder.RenameIndex(
                name: "IX_profiles_user_id",
                table: "profiles",
                newName: "idx_profiles_user_id");

            migrationBuilder.RenameIndex(
                name: "IX_refresh_tokens_token",
                table: "refresh_tokens",
                newName: "idx_refresh_tokens_token");

            migrationBuilder.RenameIndex(
                name: "IX_refresh_tokens_user_id",
                table: "refresh_tokens",
                newName: "idx_refresh_tokens_user_id");

            migrationBuilder.RenameIndex(
                name: "IX_users_email",
                table: "users",
                newName: "idx_users_email");

            migrationBuilder.RenameIndex(
                name: "IX_watchlist_items_profile_id_content_id",
                table: "watchlist_items",
                newName: "idx_watchlist_items_profile_id_content_id");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            // Reverse index renames
            migrationBuilder.RenameIndex(
                name: "idx_watchlist_items_profile_id_content_id",
                table: "watchlist_items",
                newName: "IX_watchlist_items_profile_id_content_id");

            migrationBuilder.RenameIndex(
                name: "idx_users_email",
                table: "users",
                newName: "IX_users_email");

            migrationBuilder.RenameIndex(
                name: "idx_refresh_tokens_user_id",
                table: "refresh_tokens",
                newName: "IX_refresh_tokens_user_id");

            migrationBuilder.RenameIndex(
                name: "idx_refresh_tokens_token",
                table: "refresh_tokens",
                newName: "IX_refresh_tokens_token");

            migrationBuilder.RenameIndex(
                name: "idx_profiles_user_id",
                table: "profiles",
                newName: "IX_profiles_user_id");

            // Reverse FK renames
            migrationBuilder.Sql("ALTER TABLE watchlist_items RENAME CONSTRAINT \"fk_watchlist_items_profile_id\" TO \"FK_watchlist_items_profiles_profile_id\";");
            migrationBuilder.Sql("ALTER TABLE refresh_tokens RENAME CONSTRAINT \"fk_refresh_tokens_user_id\" TO \"FK_refresh_tokens_users_user_id\";");
            migrationBuilder.Sql("ALTER TABLE profiles RENAME CONSTRAINT \"fk_profiles_user_id\" TO \"FK_profiles_users_user_id\";");

            // Reverse PK renames
            migrationBuilder.Sql("ALTER TABLE watchlist_items RENAME CONSTRAINT \"pk_watchlist_items\" TO \"PK_watchlist_items\";");
            migrationBuilder.Sql("ALTER TABLE refresh_tokens RENAME CONSTRAINT \"pk_refresh_tokens\" TO \"PK_refresh_tokens\";");
            migrationBuilder.Sql("ALTER TABLE profiles RENAME CONSTRAINT \"pk_profiles\" TO \"PK_profiles\";");
            migrationBuilder.Sql("ALTER TABLE users RENAME CONSTRAINT \"pk_users\" TO \"PK_users\";");
        }
    }
}
