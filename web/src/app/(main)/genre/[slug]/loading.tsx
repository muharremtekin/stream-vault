import { Skeleton } from '@/components/ui/skeleton';

const SKELETON_COUNT = 20;

export default function GenreLoading() {
  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      <Skeleton variant="text" className="mb-2 h-9 w-48" />
      <Skeleton variant="text" className="mb-8 h-5 w-24" />
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          // eslint-disable-next-line react/no-array-index-key
          <Skeleton key={`genre-card-${i}`} variant="custom" className="aspect-[2/3] w-full rounded-md" />
        ))}
      </div>
    </div>
  );
}
