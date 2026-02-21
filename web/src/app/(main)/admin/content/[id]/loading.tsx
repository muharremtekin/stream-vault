import { Skeleton } from '@/components/ui/skeleton';

export default function ContentDetailLoading() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Skeleton variant="custom" className="h-8 w-8 rounded-md" />
        <Skeleton variant="text" className="h-7 w-64" />
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <Skeleton variant="custom" className="h-24 w-full rounded-lg" />
          <div className="grid gap-4 sm:grid-cols-3">
            <Skeleton variant="custom" className="h-16 rounded-lg" />
            <Skeleton variant="custom" className="h-16 rounded-lg" />
            <Skeleton variant="custom" className="h-16 rounded-lg" />
          </div>
        </div>
        <Skeleton variant="card" className="h-64 w-full" />
      </div>
      <Skeleton variant="custom" className="h-40 w-full rounded-lg" />
    </div>
  );
}
