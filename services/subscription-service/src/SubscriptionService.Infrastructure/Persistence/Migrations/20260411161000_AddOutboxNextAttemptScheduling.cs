using System;
using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace SubscriptionService.Infrastructure.Persistence.Migrations
{
    [DbContext(typeof(SubscriptionDbContext))]
    [Migration("20260411161000_AddOutboxNextAttemptScheduling")]
    public partial class AddOutboxNextAttemptScheduling : Migration
    {
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<DateTime>(
                name: "next_attempt_at",
                table: "outbox_messages",
                type: "timestamp with time zone",
                nullable: true);

            migrationBuilder.Sql("""
                UPDATE outbox_messages
                SET next_attempt_at = COALESCE(last_attempted_at, created_at)
                WHERE processed_at IS NULL AND next_attempt_at IS NULL;
                """);
        }

        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "next_attempt_at",
                table: "outbox_messages");
        }
    }
}
