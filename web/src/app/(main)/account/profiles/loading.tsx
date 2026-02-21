import { Skeleton } from '@/components/ui/skeleton';

export default function ProfilesLoading() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <Skeleton variant="text" className="mb-8 h-8 w-48" />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {['prof-skel-1', 'prof-skel-2', 'prof-skel-3'].map((id) => (
          <div key={id} className="rounded-lg border border-border bg-card p-5">
            <div className="flex items-center gap-4">
              <Skeleton className="h-14 w-14 rounded-md" />
              <div className="flex-1 space-y-2">
                <Skeleton variant="text" className="h-5 w-24" />
                <Skeleton variant="text" className="h-4 w-16" />
              </div>
            </div>
            <div className="mt-4 flex gap-2">
              <Skeleton variant="text" className="h-8 w-20" />
              <Skeleton variant="text" className="h-8 w-20" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
