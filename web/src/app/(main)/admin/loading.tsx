import { Skeleton } from '@/components/ui/skeleton';

export default function AdminDashboardLoading() {
  return (
    <>
      <Skeleton variant="text" className="mb-8 h-8 w-40" />

      <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {['stat-1', 'stat-2', 'stat-3', 'stat-4'].map((id) => (
          <div key={id} className="rounded-lg border border-border bg-card p-5">
            <div className="flex items-center gap-3">
              <Skeleton variant="circle" className="h-10 w-10" />
              <div className="flex-1 space-y-1">
                <Skeleton variant="text" className="h-4 w-20" />
                <Skeleton variant="text" className="h-7 w-16" />
              </div>
            </div>
          </div>
        ))}
      </div>

      <Skeleton variant="text" className="mb-4 h-6 w-48" />
      <div className="rounded-lg border border-border">
        <div className="border-b border-border bg-muted/50 px-4 py-3">
          <div className="flex gap-8">
            <Skeleton variant="text" className="h-4 w-20" />
            <Skeleton variant="text" className="h-4 w-16" />
            <Skeleton variant="text" className="h-4 w-32" />
            <Skeleton variant="text" className="h-4 w-20" />
          </div>
        </div>
        {['row-1', 'row-2', 'row-3', 'row-4', 'row-5'].map((id) => (
          <div key={id} className="border-b border-border px-4 py-3 last:border-b-0">
            <div className="flex gap-8">
              <Skeleton variant="text" className="h-4 w-24" />
              <Skeleton variant="text" className="h-5 w-16" />
              <Skeleton variant="text" className="h-2 w-32" />
              <Skeleton variant="text" className="h-4 w-20" />
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
