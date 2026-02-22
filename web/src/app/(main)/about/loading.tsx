import { Skeleton } from '@/components/ui/skeleton';

export default function AboutLoading() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:py-12">
      {/* Hero skeleton */}
      <div className="mb-16 rounded-xl border border-border bg-card px-6 py-16 text-center sm:px-12 sm:py-20">
        <Skeleton variant="text" className="mx-auto mb-4 h-10 w-64" />
        <Skeleton variant="text" className="mx-auto h-5 w-96 max-w-full" />
      </div>

      {/* Mission skeleton */}
      <div className="mb-16 text-center">
        <Skeleton variant="text" className="mx-auto mb-4 h-8 w-48" />
        <Skeleton variant="text" className="mx-auto h-5 w-full max-w-3xl" />
        <Skeleton variant="text" className="mx-auto mt-2 h-5 w-3/4 max-w-3xl" />
      </div>

      {/* Features grid skeleton */}
      <div className="mb-16">
        <Skeleton variant="text" className="mx-auto mb-8 h-8 w-56" />
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          {['feat-skel-1', 'feat-skel-2', 'feat-skel-3', 'feat-skel-4'].map((id) => (
            <div key={id} className="rounded-lg border border-border bg-card p-6">
              <Skeleton className="mb-4 h-12 w-12 rounded-full" />
              <Skeleton variant="text" className="mb-2 h-5 w-40" />
              <Skeleton variant="text" className="h-4 w-full" />
            </div>
          ))}
        </div>
      </div>

      {/* Stats skeleton */}
      <div className="mb-16">
        <Skeleton variant="text" className="mx-auto mb-8 h-8 w-52" />
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {['stat-skel-1', 'stat-skel-2', 'stat-skel-3', 'stat-skel-4'].map((id) => (
            <div key={id} className="rounded-lg border border-border bg-card p-6 text-center">
              <Skeleton variant="text" className="mx-auto h-8 w-20" />
              <Skeleton variant="text" className="mx-auto mt-1 h-4 w-24" />
            </div>
          ))}
        </div>
      </div>

      {/* Team skeleton */}
      <div className="text-center">
        <Skeleton variant="text" className="mx-auto mb-4 h-8 w-48" />
        <Skeleton variant="text" className="mx-auto h-5 w-full max-w-3xl" />
        <Skeleton variant="text" className="mx-auto mt-2 h-5 w-2/3 max-w-3xl" />
        <Skeleton variant="text" className="mx-auto mt-6 h-10 w-56 rounded-md" />
      </div>
    </div>
  );
}
