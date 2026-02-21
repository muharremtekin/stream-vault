import { Skeleton } from '@/components/ui/skeleton';

export default function AccountLoading() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <Skeleton variant="text" className="mb-8 h-8 w-40" />

      {/* User info card skeleton */}
      <div className="mb-8 rounded-lg border border-border bg-card p-6">
        <div className="flex items-start gap-4">
          <Skeleton className="h-14 w-14 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton variant="text" className="h-6 w-48" />
            <Skeleton variant="text" className="h-5 w-32" />
          </div>
        </div>
        <div className="mt-6 border-t border-border pt-4">
          <Skeleton variant="text" className="h-5 w-56" />
        </div>
      </div>

      {/* Navigation cards skeleton */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {['nav-skel-1', 'nav-skel-2', 'nav-skel-3'].map((id) => (
          <div key={id} className="rounded-lg border border-border bg-card p-5">
            <div className="flex items-center gap-4">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="flex-1 space-y-1.5">
                <Skeleton variant="text" className="h-5 w-24" />
                <Skeleton variant="text" className="h-3 w-40" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
