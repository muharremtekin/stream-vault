import { Skeleton } from '@/components/ui/skeleton';

const SKELETON_COUNT = 8;

export default function MyListLoading() {
  return (
    <main className="min-h-screen bg-background px-4 sm:px-8 lg:px-12 pt-24 pb-16">
      {/* Page title skeleton */}
      <Skeleton variant="text" className="mb-6 h-8 w-40" />

      {/* Card grid skeleton */}
      <div className="flex flex-wrap gap-2 md:gap-3">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          <div key={`skeleton-${i}`} className="shrink-0 w-32 sm:w-36 md:w-44">
            <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
          </div>
        ))}
      </div>
    </main>
  );
}
