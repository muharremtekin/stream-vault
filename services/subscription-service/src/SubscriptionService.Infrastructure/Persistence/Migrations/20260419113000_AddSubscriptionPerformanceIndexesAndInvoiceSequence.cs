using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace SubscriptionService.Infrastructure.Persistence.Migrations
{
    [DbContext(typeof(SubscriptionDbContext))]
    [Migration("20260419113000_AddSubscriptionPerformanceIndexesAndInvoiceSequence")]
    public partial class AddSubscriptionPerformanceIndexesAndInvoiceSequence : Migration
    {
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateSequence<long>(
                name: "invoice_numbers");

            migrationBuilder.CreateIndex(
                name: "idx_invoices_subscription_issued_at",
                table: "invoices",
                columns: new[] { "subscription_id", "issued_at" });

            migrationBuilder.CreateIndex(
                name: "idx_payments_subscription_created_at",
                table: "payments",
                columns: new[] { "subscription_id", "created_at" });

            migrationBuilder.CreateIndex(
                name: "idx_subscriptions_status_period_end",
                table: "subscriptions",
                columns: new[] { "status", "period_end" });

            migrationBuilder.CreateIndex(
                name: "idx_subscriptions_user_status",
                table: "subscriptions",
                columns: new[] { "user_id", "status" });
        }

        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropIndex(
                name: "idx_invoices_subscription_issued_at",
                table: "invoices");

            migrationBuilder.DropIndex(
                name: "idx_payments_subscription_created_at",
                table: "payments");

            migrationBuilder.DropIndex(
                name: "idx_subscriptions_status_period_end",
                table: "subscriptions");

            migrationBuilder.DropIndex(
                name: "idx_subscriptions_user_status",
                table: "subscriptions");

            migrationBuilder.DropSequence(
                name: "invoice_numbers");
        }
    }
}
