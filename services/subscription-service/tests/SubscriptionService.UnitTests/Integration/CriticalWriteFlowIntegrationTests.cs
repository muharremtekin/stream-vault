using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Storage;
using Microsoft.Extensions.Logging.Abstractions;
using SubscriptionService.Application.Commands.CancelSubscription;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Application.Interfaces;
using SubscriptionService.Application.Sagas;
using SubscriptionService.Domain.Entities;
using SubscriptionService.Domain.Enums;
using SubscriptionService.Domain.Exceptions;
using SubscriptionService.Infrastructure.Persistence;
using SubscriptionService.Infrastructure.Persistence.Repositories;
using Xunit;

namespace SubscriptionService.UnitTests.Integration;

public class CriticalWriteFlowIntegrationTests
{
    [Fact]
    public async Task CreateSubscription_CommitsPendingAndFinalPhaseSeparately()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var options = CreateOptions(dbName, root);

        await using var context = new SubscriptionDbContext(options);
        var plan = CreatePlan(Guid.NewGuid(), "Standard", PlanTier.Standard, 79.99m);
        await context.Plans.AddAsync(plan);
        await context.SaveChangesAsync();

        var paymentGateway = new TestPaymentGateway(
            new PaymentResult(true, "txn_create_success", null));
        var unitOfWork = new RecordingSubscriptionUnitOfWork(
            context,
            saveCount => CaptureSnapshot(options, saveCount));

        var saga = CreateSubscriptionSaga(
            context,
            unitOfWork,
            paymentGateway);

        var result = await saga.ExecuteAsync(Guid.NewGuid(), plan.Id, "4242424242424242", CancellationToken.None);

        result.Status.Should().Be(SubscriptionStatus.Active);
        unitOfWork.SaveCallCount.Should().Be(2);

        unitOfWork.Snapshots[0].SubscriptionStatus.Should().Be(SubscriptionStatus.PendingPayment);
        unitOfWork.Snapshots[0].Payments.Should().Be(0);
        unitOfWork.Snapshots[0].Invoices.Should().Be(0);
        unitOfWork.Snapshots[0].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[0].SagaStatus.Should().Be(SagaStatus.InProgress);

        unitOfWork.Snapshots[1].SubscriptionStatus.Should().Be(SubscriptionStatus.Active);
        unitOfWork.Snapshots[1].Payments.Should().Be(1);
        unitOfWork.Snapshots[1].Invoices.Should().Be(1);
        unitOfWork.Snapshots[1].OutboxMessages.Should().Be(2);
        unitOfWork.Snapshots[1].SagaStatus.Should().Be(SagaStatus.Completed);
    }

    [Fact]
    public async Task CreateSubscription_WhenInvoiceGenerationFails_RefundsAndLeavesConsistentState()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var options = CreateOptions(dbName, root);

        await using var context = new SubscriptionDbContext(options);
        var plan = CreatePlan(Guid.NewGuid(), "Premium", PlanTier.Premium, 119.99m);
        await context.Plans.AddAsync(plan);
        await context.SaveChangesAsync();

        var paymentGateway = new TestPaymentGateway(
            new PaymentResult(true, "txn_invoice_failure", null));
        var unitOfWork = new RecordingSubscriptionUnitOfWork(
            context,
            saveCount => CaptureSnapshot(options, saveCount));

        var saga = CreateSubscriptionSaga(
            context,
            unitOfWork,
            paymentGateway,
            new ThrowingInvoiceRepository(new InvoiceRepository(context)));

        var act = () => saga.ExecuteAsync(Guid.NewGuid(), plan.Id, "4242424242424242", CancellationToken.None);

        await act.Should().ThrowAsync<PaymentFailedException>();

        paymentGateway.Refunds.Should().ContainSingle(r =>
            r.TransactionId == "txn_invoice_failure" && r.Amount == 0);
        unitOfWork.SaveCallCount.Should().Be(3);

        unitOfWork.Snapshots[0].SubscriptionStatus.Should().Be(SubscriptionStatus.PendingPayment);
        unitOfWork.Snapshots[0].OutboxMessages.Should().Be(0);

        unitOfWork.Snapshots[1].SubscriptionStatus.Should().Be(SubscriptionStatus.PendingPayment);
        unitOfWork.Snapshots[1].Payments.Should().Be(0);
        unitOfWork.Snapshots[1].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[1].SagaStatus.Should().Be(SagaStatus.Compensating);

        unitOfWork.Snapshots[2].SubscriptionStatus.Should().Be(SubscriptionStatus.Failed);
        unitOfWork.Snapshots[2].Payments.Should().Be(0);
        unitOfWork.Snapshots[2].Invoices.Should().Be(0);
        unitOfWork.Snapshots[2].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[2].SagaStatus.Should().Be(SagaStatus.Failed);
    }

    [Fact]
    public async Task ChangePlan_Upgrade_PersistsPaymentSubscriptionAndOutboxInFinalCommit()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var options = CreateOptions(dbName, root);

        await using var context = new SubscriptionDbContext(options);
        var standardPlan = CreatePlan(Guid.NewGuid(), "Standard", PlanTier.Standard, 79.99m);
        var premiumPlan = CreatePlan(Guid.NewGuid(), "Premium", PlanTier.Premium, 119.99m);
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = Guid.NewGuid(),
            PlanId = standardPlan.Id,
            Plan = standardPlan,
            Status = SubscriptionStatus.Active,
            PeriodStart = DateTime.UtcNow.AddDays(-5),
            PeriodEnd = DateTime.UtcNow.AddDays(25),
            AutoRenew = true
        };

        await context.Plans.AddRangeAsync(standardPlan, premiumPlan);
        await context.Subscriptions.AddAsync(subscription);
        await context.SaveChangesAsync();

        var paymentGateway = new TestPaymentGateway(
            new PaymentResult(true, "txn_change_plan", null));
        var unitOfWork = new RecordingSubscriptionUnitOfWork(
            context,
            saveCount => CaptureSnapshot(options, saveCount, subscription.Id));

        var saga = CreateChangePlanSaga(
            context,
            unitOfWork,
            paymentGateway);

        var result = await saga.ExecuteAsync(
            subscription.Id,
            premiumPlan.Id,
            "4242424242424242",
            CancellationToken.None);

        result.PlanId.Should().Be(premiumPlan.Id);
        unitOfWork.SaveCallCount.Should().Be(2);

        unitOfWork.Snapshots[0].SubscriptionPlanId.Should().Be(standardPlan.Id);
        unitOfWork.Snapshots[0].Payments.Should().Be(0);
        unitOfWork.Snapshots[0].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[0].SagaStatus.Should().Be(SagaStatus.InProgress);

        unitOfWork.Snapshots[1].SubscriptionPlanId.Should().Be(premiumPlan.Id);
        unitOfWork.Snapshots[1].Payments.Should().Be(1);
        unitOfWork.Snapshots[1].OutboxMessages.Should().Be(2);
        unitOfWork.Snapshots[1].SagaStatus.Should().Be(SagaStatus.Completed);
    }

    [Fact]
    public async Task ChangePlan_WhenFinalizePhaseFails_DoesNotLeakPendingWritesIntoCompensation()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var options = CreateOptions(dbName, root);

        await using var context = new SubscriptionDbContext(options);
        var standardPlan = CreatePlan(Guid.NewGuid(), "Standard", PlanTier.Standard, 79.99m);
        var premiumPlan = CreatePlan(Guid.NewGuid(), "Premium", PlanTier.Premium, 119.99m);
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = Guid.NewGuid(),
            PlanId = standardPlan.Id,
            Plan = standardPlan,
            Status = SubscriptionStatus.Active,
            PeriodStart = DateTime.UtcNow.AddDays(-5),
            PeriodEnd = DateTime.UtcNow.AddDays(25),
            AutoRenew = true
        };

        await context.Plans.AddRangeAsync(standardPlan, premiumPlan);
        await context.Subscriptions.AddAsync(subscription);
        await context.SaveChangesAsync();

        var paymentGateway = new TestPaymentGateway(
            new PaymentResult(true, "txn_change_plan_finalize_failure", null));
        var unitOfWork = new RecordingSubscriptionUnitOfWork(
            context,
            saveCount => CaptureSnapshot(options, saveCount, subscription.Id));

        var saga = CreateChangePlanSaga(
            context,
            unitOfWork,
            paymentGateway,
            new ThrowingOutboxRepository(new OutboxRepository(context), throwOnCall: 2));

        var act = () => saga.ExecuteAsync(
            subscription.Id,
            premiumPlan.Id,
            "4242424242424242",
            CancellationToken.None);

        await act.Should().ThrowAsync<PaymentFailedException>();

        paymentGateway.Refunds.Should().ContainSingle(r =>
            r.TransactionId == "txn_change_plan_finalize_failure" && r.Amount > 0);
        unitOfWork.SaveCallCount.Should().Be(3);

        unitOfWork.Snapshots[0].SubscriptionPlanId.Should().Be(standardPlan.Id);
        unitOfWork.Snapshots[0].Payments.Should().Be(0);
        unitOfWork.Snapshots[0].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[0].SagaStatus.Should().Be(SagaStatus.InProgress);

        unitOfWork.Snapshots[1].SubscriptionPlanId.Should().Be(standardPlan.Id);
        unitOfWork.Snapshots[1].Payments.Should().Be(0);
        unitOfWork.Snapshots[1].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[1].SagaStatus.Should().Be(SagaStatus.Compensating);

        unitOfWork.Snapshots[2].SubscriptionPlanId.Should().Be(standardPlan.Id);
        unitOfWork.Snapshots[2].Payments.Should().Be(0);
        unitOfWork.Snapshots[2].OutboxMessages.Should().Be(0);
        unitOfWork.Snapshots[2].SagaStatus.Should().Be(SagaStatus.Failed);
    }

    [Fact]
    public async Task CancelSubscription_CommitsDomainChangeAndOutboxTogether()
    {
        var root = new InMemoryDatabaseRoot();
        var dbName = Guid.NewGuid().ToString("N");
        var options = CreateOptions(dbName, root);

        await using var context = new SubscriptionDbContext(options);
        var plan = CreatePlan(Guid.NewGuid(), "Basic", PlanTier.Basic, 49.99m);
        var subscription = new Subscription
        {
            Id = Guid.NewGuid(),
            UserId = Guid.NewGuid(),
            PlanId = plan.Id,
            Plan = plan,
            Status = SubscriptionStatus.Active,
            PeriodStart = DateTime.UtcNow.AddDays(-5),
            PeriodEnd = DateTime.UtcNow.AddDays(25),
            AutoRenew = true
        };

        await context.Plans.AddAsync(plan);
        await context.Subscriptions.AddAsync(subscription);
        await context.SaveChangesAsync();

        var unitOfWork = new RecordingSubscriptionUnitOfWork(
            context,
            saveCount => CaptureSnapshot(options, saveCount, subscription.Id));

        var handler = new CancelSubscriptionHandler(
            new SubscriptionRepository(context),
            new OutboxRepository(context),
            unitOfWork,
            NullLogger<CancelSubscriptionHandler>.Instance);

        var result = await handler.Handle(
            new CancelSubscriptionCommand(subscription.UserId),
            CancellationToken.None);

        result.Status.Should().Be("Cancelled");
        unitOfWork.SaveCallCount.Should().Be(1);
        unitOfWork.Snapshots[0].SubscriptionStatus.Should().Be(SubscriptionStatus.Cancelled);
        unitOfWork.Snapshots[0].OutboxMessages.Should().Be(1);
    }

    private static SubscriptionSaga CreateSubscriptionSaga(
        SubscriptionDbContext context,
        RecordingSubscriptionUnitOfWork unitOfWork,
        IPaymentGateway paymentGateway,
        IInvoiceRepository? invoiceRepository = null)
    {
        var subscriptionRepository = new SubscriptionRepository(context);
        var sagaRepository = new SagaRepository(context);

        var compensatingActions = new CompensatingActions(
            subscriptionRepository,
            paymentGateway,
            sagaRepository,
            unitOfWork,
            NullLogger<CompensatingActions>.Instance);

        return new SubscriptionSaga(
            new PlanRepository(context),
            subscriptionRepository,
            new PaymentRepository(context),
            paymentGateway,
            invoiceRepository ?? new InvoiceRepository(context),
            new OutboxRepository(context),
            sagaRepository,
            unitOfWork,
            compensatingActions,
            NullLogger<SubscriptionSaga>.Instance);
    }

    private static ChangePlanSaga CreateChangePlanSaga(
        SubscriptionDbContext context,
        RecordingSubscriptionUnitOfWork unitOfWork,
        IPaymentGateway paymentGateway,
        IOutboxRepository? outboxRepository = null)
    {
        var subscriptionRepository = new SubscriptionRepository(context);
        var sagaRepository = new SagaRepository(context);

        var compensatingActions = new CompensatingActions(
            subscriptionRepository,
            paymentGateway,
            sagaRepository,
            unitOfWork,
            NullLogger<CompensatingActions>.Instance);

        return new ChangePlanSaga(
            new PlanRepository(context),
            subscriptionRepository,
            new PaymentRepository(context),
            paymentGateway,
            outboxRepository ?? new OutboxRepository(context),
            sagaRepository,
            unitOfWork,
            compensatingActions,
            NullLogger<ChangePlanSaga>.Instance);
    }

    private static DbContextOptions<SubscriptionDbContext> CreateOptions(string dbName, InMemoryDatabaseRoot root)
    {
        return new DbContextOptionsBuilder<SubscriptionDbContext>()
            .UseInMemoryDatabase(dbName, root)
            .Options;
    }

    private static WriteSnapshot CaptureSnapshot(
        DbContextOptions<SubscriptionDbContext> options,
        int saveCount,
        Guid? subscriptionId = null)
    {
        using var snapshotContext = new SubscriptionDbContext(options);

        var subscription = subscriptionId.HasValue
            ? snapshotContext.Subscriptions.FirstOrDefault(s => s.Id == subscriptionId.Value)
            : snapshotContext.Subscriptions.OrderBy(s => s.CreatedAt).FirstOrDefault();
        var saga = snapshotContext.SagaStates.OrderBy(s => s.CreatedAt).FirstOrDefault();

        return new WriteSnapshot(
            saveCount,
            subscription?.Status,
            subscription?.PlanId,
            snapshotContext.Payments.Count(),
            snapshotContext.Invoices.Count(),
            snapshotContext.OutboxMessages.Count(),
            saga?.Status);
    }

    private static Plan CreatePlan(Guid id, string name, PlanTier tier, decimal price)
    {
        return new Plan
        {
            Id = id,
            Name = name,
            Tier = tier,
            PriceMonthly = price,
            MaxScreens = tier switch
            {
                PlanTier.Basic => 1,
                PlanTier.Standard => 3,
                _ => 4
            },
            MaxQuality = tier == PlanTier.Premium ? "4K" : "1080p",
            Features = "Test features",
            IsActive = true
        };
    }

    private sealed record WriteSnapshot(
        int SaveCount,
        SubscriptionStatus? SubscriptionStatus,
        Guid? SubscriptionPlanId,
        int Payments,
        int Invoices,
        int OutboxMessages,
        SagaStatus? SagaStatus);

    private sealed class RecordingSubscriptionUnitOfWork : ISubscriptionUnitOfWork
    {
        private readonly SubscriptionDbContext _context;
        private readonly Func<int, WriteSnapshot> _snapshotFactory;

        public RecordingSubscriptionUnitOfWork(
            SubscriptionDbContext context,
            Func<int, WriteSnapshot> snapshotFactory)
        {
            _context = context;
            _snapshotFactory = snapshotFactory;
        }

        public int SaveCallCount { get; private set; }
        public List<WriteSnapshot> Snapshots { get; } = new();

        public async Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
        {
            var result = await _context.SaveChangesAsync(cancellationToken);
            SaveCallCount++;
            Snapshots.Add(_snapshotFactory(SaveCallCount));
            return result;
        }

        public void DiscardPendingChanges()
        {
            _context.ChangeTracker.Clear();
        }
    }

    private sealed class TestPaymentGateway : IPaymentGateway
    {
        private readonly Queue<PaymentResult> _chargeResults;

        public TestPaymentGateway(params PaymentResult[] chargeResults)
        {
            _chargeResults = new Queue<PaymentResult>(chargeResults);
        }

        public List<(string TransactionId, decimal Amount)> Refunds { get; } = new();

        public Task<PaymentResult> ChargeAsync(
            string cardNumber,
            decimal amount,
            string currency,
            CancellationToken cancellationToken = default)
        {
            return Task.FromResult(_chargeResults.Dequeue());
        }

        public Task<PaymentResult> RefundAsync(
            string transactionId,
            decimal amount,
            CancellationToken cancellationToken = default)
        {
            Refunds.Add((transactionId, amount));
            return Task.FromResult(new PaymentResult(true, $"refund_{transactionId}", null));
        }
    }

    private sealed class ThrowingInvoiceRepository : IInvoiceRepository
    {
        private readonly IInvoiceRepository _inner;

        public ThrowingInvoiceRepository(IInvoiceRepository inner)
        {
            _inner = inner;
        }

        public Task<Invoice?> GetByIdAsync(Guid id, CancellationToken cancellationToken = default)
            => _inner.GetByIdAsync(id, cancellationToken);

        public Task AddAsync(Invoice invoice, CancellationToken cancellationToken = default)
            => _inner.AddAsync(invoice, cancellationToken);

        public Task<List<Invoice>> GetBySubscriptionIdAsync(
            Guid subscriptionId,
            CancellationToken cancellationToken = default)
            => _inner.GetBySubscriptionIdAsync(subscriptionId, cancellationToken);

        public Task<List<InvoiceDto>> GetDtosBySubscriptionIdAsync(
            Guid subscriptionId,
            CancellationToken cancellationToken = default)
            => _inner.GetDtosBySubscriptionIdAsync(subscriptionId, cancellationToken);

        public Task<string> GenerateInvoiceNumberAsync(CancellationToken cancellationToken = default)
            => throw new InvalidOperationException("DB error");
    }

    private sealed class ThrowingOutboxRepository : IOutboxRepository
    {
        private readonly IOutboxRepository _inner;
        private readonly int _throwOnCall;
        private int _addCallCount;

        public ThrowingOutboxRepository(IOutboxRepository inner, int throwOnCall)
        {
            _inner = inner;
            _throwOnCall = throwOnCall;
        }

        public Task AddAsync(OutboxMessage message, CancellationToken cancellationToken = default)
        {
            _addCallCount++;
            if (_addCallCount == _throwOnCall)
            {
                throw new InvalidOperationException("DB error");
            }

            return _inner.AddAsync(message, cancellationToken);
        }

        public Task<List<OutboxMessage>> GetUnprocessedAsync(
            int batchSize = 50,
            CancellationToken cancellationToken = default)
            => _inner.GetUnprocessedAsync(batchSize, cancellationToken);

        public Task MarkAsProcessedAsync(Guid id, CancellationToken cancellationToken = default)
            => _inner.MarkAsProcessedAsync(id, cancellationToken);

        public Task IncrementRetryAsync(
            Guid id,
            string errorMessage,
            CancellationToken cancellationToken = default)
            => _inner.IncrementRetryAsync(id, errorMessage, cancellationToken);

        public Task MarkAsDeadLetterAsync(Guid id, CancellationToken cancellationToken = default)
            => _inner.MarkAsDeadLetterAsync(id, cancellationToken);

        public Task<List<OutboxMessage>> GetDeadLetterMessagesAsync(
            int batchSize = 50,
            CancellationToken cancellationToken = default)
            => _inner.GetDeadLetterMessagesAsync(batchSize, cancellationToken);
    }
}
