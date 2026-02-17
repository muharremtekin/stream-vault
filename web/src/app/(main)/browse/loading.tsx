import { Skeleton } from '@/components/ui/skeleton';

const ROW_COUNT = 3;
const CARD_COUNT = 6;

export default function BrowseLoading() {
  return (
    <div className="min-h-screen bg-background">
      {/* Hero skeleton */}
      <Skeleton variant="custom" className="h-[80vh] min-h-[500px] w-full rounded-none" />

      {/* Content row skeletons */}
      <div className="space-y-8 py-4 -mt-16 relative z-10">
        {Array.from({ length: ROW_COUNT }).map((_, rowIdx) => (
          // eslint-disable-next-line react/no-array-index-key
          <div key={`row-${rowIdx}`} className="px-4 sm:px-8 lg:px-12">
            <Skeleton variant="text" className="mb-3 h-6 w-48" />
            <div className="flex gap-2">
              {Array.from({ length: CARD_COUNT }).map((__, cardIdx) => (
                // eslint-disable-next-line react/no-array-index-key
                <Skeleton
                  key={`card-${rowIdx}-${cardIdx}`}
                  variant="card"
                  className="shrink-0 w-32 sm:w-36 md:w-44 aspect-[2/3] rounded-md"
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
