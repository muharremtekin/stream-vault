using System.Collections.Concurrent;
using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Storage;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Infrastructure.Persistence;
using SubscriptionService.Infrastructure.Persistence.Repositories;
using Xunit;

namespace SubscriptionService.UnitTests.Integration;

public class InvoiceRepositoryTests
{
    [Fact]
    public async Task GenerateInvoiceNumberAsync_WithNonRelationalProvider_ReturnsCurrentFormat()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");

        await using var context = CreateOptions(dbName, root);
        var repository = new InvoiceRepository(context);

        var invoiceNumber = await repository.GenerateInvoiceNumberAsync();

        invoiceNumber.Should().MatchRegex(@"^INV-\d{8}-\d{6}$");
    }

    [Fact]
    public async Task GenerateInvoiceNumberAsync_WithConcurrentNonRelationalCalls_ReturnsUniqueNumbers()
    {
        const int workerCount = 8;
        const int invoicesPerWorker = 128;
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var generatedInvoiceNumbers = new ConcurrentBag<string>();
        using var barrier = new Barrier(workerCount);

        var workers = Enumerable.Range(0, workerCount).Select(_ => Task.Factory.StartNew(
            () =>
            {
                using var context = CreateOptions(dbName, root);
                var repository = new InvoiceRepository(context);
                // Initialize each context before synchronizing the invoice requests.
                context.Database.IsRelational();

                for (var index = 0; index < invoicesPerWorker; index++)
                {
                    // Dedicated threads avoid thread-pool starvation at the barrier.
                    if (!barrier.SignalAndWait(TimeSpan.FromSeconds(30)))
                        throw new TimeoutException("Invoice workers did not reach the barrier in time.");

                    generatedInvoiceNumbers.Add(repository.GenerateInvoiceNumberAsync().GetAwaiter().GetResult());
                }
            },
            CancellationToken.None,
            TaskCreationOptions.LongRunning,
            TaskScheduler.Default)).ToArray();

        await Task.WhenAll(workers);

        generatedInvoiceNumbers.Should().HaveCount(workerCount * invoicesPerWorker);
        generatedInvoiceNumbers.Should().OnlyHaveUniqueItems();
    }

    [Fact]
    public async Task GenerateInvoiceNumberAsync_WithNonRelationalProvider_DoesNotDependOnInvoiceCount()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");

        await using var context = CreateOptions(dbName, root);
        var repository = new InvoiceRepository(context);

        var firstInvoiceNumber = await repository.GenerateInvoiceNumberAsync();

        await context.Invoices.AddRangeAsync(Enumerable.Range(0, 25).Select(index => new Invoice
        {
            SubscriptionId = Guid.NewGuid(),
            InvoiceNumber = $"SEEDED-{index}",
            Amount = 79.99m,
            Currency = "TRY",
            PeriodStart = DateTime.UtcNow.Date,
            PeriodEnd = DateTime.UtcNow.Date.AddDays(30),
            IssuedAt = DateTime.UtcNow
        }));
        await context.SaveChangesAsync();

        var secondInvoiceNumber = await repository.GenerateInvoiceNumberAsync();

        ParseSuffix(secondInvoiceNumber).Should().Be(ParseSuffix(firstInvoiceNumber) + 1);
    }

    private static SubscriptionDbContext CreateOptions(string dbName, InMemoryDatabaseRoot root)
    {
        var options = new DbContextOptionsBuilder<SubscriptionDbContext>()
            .UseInMemoryDatabase(dbName, root)
            .Options;

        return new SubscriptionDbContext(options);
    }

    private static int ParseSuffix(string invoiceNumber)
        => int.Parse(invoiceNumber.Split('-')[2]);
}
