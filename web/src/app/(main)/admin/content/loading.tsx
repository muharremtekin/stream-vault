import { Skeleton } from '@/components/ui/skeleton';

export default function ContentManagementLoading() {
  return (
    <>
      <Skeleton variant="text" className="mb-8 h-8 w-52" />

      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Skeleton variant="text" className="h-10 w-full sm:max-w-xs" />
        <Skeleton variant="text" className="h-8 w-24" />
      </div>

      <Skeleton variant="text" className="mb-2 h-4 w-16" />
      <div className="rounded-lg border border-border">
        <div className="border-b border-border bg-muted/50 px-4 py-3">
          <div className="flex gap-12">
            <Skeleton variant="text" className="h-4 w-20" />
            <Skeleton variant="text" className="h-4 w-12" />
            <Skeleton variant="text" className="h-4 w-14" />
            <Skeleton variant="text" className="h-4 w-12" />
          </div>
        </div>
        {['cl-1', 'cl-2', 'cl-3', 'cl-4', 'cl-5', 'cl-6', 'cl-7', 'cl-8'].map((id) => (
          <div key={id} className="border-b border-border px-4 py-3 last:border-b-0">
            <div className="flex gap-12">
              <Skeleton variant="text" className="h-4 w-40" />
              <Skeleton variant="text" className="h-5 w-14" />
              <Skeleton variant="text" className="h-5 w-16" />
              <Skeleton variant="text" className="h-4 w-20" />
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
