import { Skeleton } from '@/components/ui/skeleton';

const FILTER_CHIP_COUNT = 6;
const SKELETON_COUNT = 24;

export default function MoviesLoading() {
  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      {/* Title skeleton */}
      <Skeleton variant="text" className="mb-6 h-9 w-36" />

      {/* Filter bar skeleton */}
      <div className="mb-8 space-y-4">
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: FILTER_CHIP_COUNT }).map((_, i) => (
            <Skeleton
              key={`filter-skeleton-${String(i)}`}
              variant="custom"
              className="h-7 w-20 rounded-full"
            />
          ))}
        </div>
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton
              key={`sort-skeleton-${String(i)}`}
              variant="custom"
              className="h-7 w-16 rounded-full"
            />
          ))}
        </div>
      </div>

      {/* Grid skeleton */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          <Skeleton
            key={`movie-skeleton-${String(i)}`}
            variant="custom"
            className="aspect-[2/3] w-full rounded-md"
          />
        ))}
      </div>
    </div>
  );
}
