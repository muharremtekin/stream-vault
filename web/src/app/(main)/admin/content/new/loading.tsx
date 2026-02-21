import { Skeleton } from '@/components/ui/skeleton';

export default function ContentNewLoading() {
  return (
    <div>
      <Skeleton variant="text" className="mb-6 h-8 w-52" />
      <div className="rounded-lg border border-border bg-card p-6 space-y-6">
        <div className="flex gap-2">
          <Skeleton variant="text" className="h-10 w-24 rounded-lg" />
          <Skeleton variant="text" className="h-10 w-24 rounded-lg" />
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <Skeleton variant="text" className="h-10 w-full" />
          <Skeleton variant="text" className="h-10 w-full" />
        </div>
        <Skeleton variant="custom" className="h-24 w-full rounded-md" />
        <div className="grid gap-4 sm:grid-cols-3">
          <Skeleton variant="text" className="h-10 w-full" />
          <Skeleton variant="text" className="h-10 w-full" />
          <Skeleton variant="text" className="h-10 w-full" />
        </div>
        <Skeleton variant="text" className="h-10 w-full" />
        <div className="flex gap-3">
          <Skeleton variant="text" className="h-10 w-20" />
          <Skeleton variant="text" className="h-10 w-20" />
        </div>
      </div>
    </div>
  );
}
