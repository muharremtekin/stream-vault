using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace SubscriptionService.Infrastructure.Persistence.Migrations
{
    [DbContext(typeof(SubscriptionDbContext))]
    [Migration("20260411164500_AddOutboxPollingIndex")]
    public partial class AddOutboxPollingIndex : Migration
    {
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateIndex(
                name: "idx_outbox_messages_due_polling",
                table: "outbox_messages",
                columns: new[] { "next_attempt_at", "created_at" },
                filter: "\"processed_at\" IS NULL AND \"is_dead_letter\" = false");
        }

        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropIndex(
                name: "idx_outbox_messages_due_polling",
                table: "outbox_messages");
        }
    }
}
