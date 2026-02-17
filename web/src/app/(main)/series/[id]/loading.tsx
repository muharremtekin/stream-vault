import { Skeleton } from '@/components/ui/skeleton';

export default function SeriesLoading() {
  return (
    <>
      {/* Hero skeleton */}
      <div className="relative min-h-[70vh] w-full bg-muted">
        <Skeleton variant="custom" className="absolute inset-0" />
        <div className="absolute bottom-0 left-0 right-0 px-4 pb-12 sm:px-8 lg:px-12 sm:pb-16 space-y-4">
          <Skeleton variant="text" className="h-5 w-20" />
          <Skeleton variant="text" className="h-12 w-2/3" />
          <Skeleton variant="text" className="h-4 w-52" />
          <Skeleton variant="text" className="h-4 w-36" />
          <div className="flex gap-3 pt-2">
            <Skeleton variant="custom" className="h-11 w-28 rounded-md" />
            <Skeleton variant="custom" className="h-11 w-36 rounded-md" />
          </div>
        </div>
      </div>

      {/* Info + seasons skeleton */}
      <div className="px-4 sm:px-8 lg:px-12 py-10 space-y-10">
        <div className="space-y-4">
          <Skeleton variant="text" className="h-7 w-32" />
          <Skeleton variant="text" className="h-4 w-full" />
          <Skeleton variant="text" className="h-4 w-5/6" />
        </div>

        {/* Season tabs skeleton */}
        <div className="space-y-4">
          <Skeleton variant="text" className="h-7 w-36" />
          <div className="flex gap-2">
            {Array.from({ length: 3 }).map((_, i) => (
              // eslint-disable-next-line react/no-array-index-key
              <Skeleton key={`season-tab-${i}`} variant="custom" className="h-14 w-24 rounded-md" />
            ))}
          </div>
          {/* Episodes skeleton */}
          {Array.from({ length: 4 }).map((_, i) => (
            // eslint-disable-next-line react/no-array-index-key
            <Skeleton key={`ep-${i}`} variant="custom" className="h-24 w-full rounded-md" />
          ))}
        </div>
      </div>
    </>
  );
}
