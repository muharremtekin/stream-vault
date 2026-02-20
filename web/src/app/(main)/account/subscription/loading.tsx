import { Skeleton } from '@/components/ui/skeleton';

export default function SubscriptionLoading() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <Skeleton variant="text" className="mb-8 h-8 w-56" />

      {/* Status card skeleton */}
      <div className="mb-8 rounded-lg border border-border bg-card p-6">
        <Skeleton variant="text" className="mb-2 h-6 w-40" />
        <Skeleton variant="text" className="mb-4 h-8 w-32" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Skeleton variant="text" className="h-12 w-full" />
          <Skeleton variant="text" className="h-12 w-full" />
          <Skeleton variant="text" className="h-12 w-full" />
        </div>
      </div>

      {/* Plan cards skeleton */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        {['skeleton-loading-1', 'skeleton-loading-2', 'skeleton-loading-3'].map((id) => (
          <div key={id} className="rounded-lg border border-border bg-card p-6">
            <Skeleton variant="text" className="mb-4 h-6 w-24" />
            <Skeleton variant="text" className="mb-6 h-10 w-32" />
            <div className="space-y-3">
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-3/4" />
            </div>
            <Skeleton variant="text" className="mt-6 h-10 w-full" />
          </div>
        ))}
      </div>
    </div>
  );
}
