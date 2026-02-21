import { Skeleton } from '@/components/ui/skeleton';

export default function EncodingLoading() {
  return (
    <>
      <Skeleton variant="text" className="mb-6 h-8 w-48" />

      <div className="mb-4 flex flex-wrap gap-2">
        {['ef-1', 'ef-2', 'ef-3', 'ef-4', 'ef-5'].map((id) => (
          <Skeleton key={id} variant="text" className="h-7 w-20 rounded-full" />
        ))}
      </div>

      <div className="rounded-lg border border-border">
        <div className="border-b border-border bg-muted/50 px-4 py-3">
          <div className="flex gap-8">
            <Skeleton variant="text" className="h-4 w-20" />
            <Skeleton variant="text" className="h-4 w-16" />
            <Skeleton variant="text" className="h-4 w-32" />
            <Skeleton variant="text" className="h-4 w-20" />
            <Skeleton variant="text" className="h-4 w-24" />
          </div>
        </div>
        {['ej-l-1', 'ej-l-2', 'ej-l-3', 'ej-l-4', 'ej-l-5'].map((id) => (
          <div key={id} className="border-b border-border px-4 py-3 last:border-b-0">
            <div className="flex gap-8">
              <Skeleton variant="text" className="h-4 w-24" />
              <Skeleton variant="text" className="h-5 w-16" />
              <Skeleton variant="text" className="h-2 w-32" />
              <Skeleton variant="text" className="h-4 w-20" />
              <Skeleton variant="text" className="h-4 w-24" />
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
