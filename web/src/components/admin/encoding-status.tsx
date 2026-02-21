'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

import { useAdminEncodingJobs } from '@/lib/hooks/use-admin';
import { Badge } from '@/components/ui/badge';
import { ProgressBar } from '@/components/ui/progress-bar';
import { Skeleton } from '@/components/ui/skeleton';
import { EncodingProgress } from '@/components/admin/encoding-progress';
import { cn } from '@/lib/utils/cn';
import { formatDateTime } from '@/lib/utils/format';

import type { EncodingJob } from '@/lib/types/encoding';
import { ENCODING_STATUS_BADGE, ENCODING_PROGRESS_VARIANT } from '@/lib/types/encoding';

const STATUS_FILTERS = ['all', 'pending', 'processing', 'completed', 'failed'] as const;
type StatusFilter = (typeof STATUS_FILTERS)[number];

export function EncodingStatus() {
  const t = useTranslations('admin.encoding');
  const [filter, setFilter] = useState<StatusFilter>('all');
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);

  const params = filter === 'all' ? { limit: 50 } : { status: filter, limit: 50 };
  const { data, isLoading } = useAdminEncodingJobs(params);
  const jobs = data?.jobs ?? [];

  if (selectedJobId) {
    return <EncodingProgress jobId={selectedJobId} onBack={() => setSelectedJobId(null)} />;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {STATUS_FILTERS.map((s) => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={cn(
              'rounded-full px-3 py-1 text-xs font-medium transition-colors',
              filter === s
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-muted/80'
            )}
          >
            {t(s)}
          </button>
        ))}
      </div>

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('contentId')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('status')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('progress')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('currentStep')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('createdAt')}</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              ['ej-sk-1', 'ej-sk-2', 'ej-sk-3', 'ej-sk-4', 'ej-sk-5'].map((id) => <JobSkeleton key={id} />)
            ) : jobs.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                  {t('noJobs')}
                </td>
              </tr>
            ) : (
              jobs.map((job) => <JobRow key={job.jobId} job={job} onClick={() => setSelectedJobId(job.jobId)} t={t} />)
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

interface JobRowProps {
  job: EncodingJob;
  onClick: () => void;
  t: (key: string) => string;
}

function JobRow({ job, onClick, t }: JobRowProps) {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick();
    }
  };

  return (
    <tr
      tabIndex={0}
      onClick={onClick}
      onKeyDown={handleKeyDown}
      className="cursor-pointer border-b border-border transition-colors last:border-b-0 hover:bg-muted/30 focus-visible:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
    >
      <td className="px-4 py-3 font-mono text-xs text-foreground" title={job.contentId}>
        {job.contentId.length > 12 ? `${job.contentId.slice(0, 12)}...` : job.contentId}
      </td>
      <td className="px-4 py-3">
        <Badge variant={ENCODING_STATUS_BADGE[job.status] ?? 'default'} size="sm">{t(job.status)}</Badge>
      </td>
      <td className="px-4 py-3 min-w-32">
        <ProgressBar value={job.progressPercentage} variant={ENCODING_PROGRESS_VARIANT[job.status] ?? 'primary'} size="sm" />
      </td>
      <td className="px-4 py-3 text-xs text-muted-foreground">{job.currentStep || '—'}</td>
      <td className="px-4 py-3 text-xs text-muted-foreground">{formatDateTime(job.createdAt)}</td>
    </tr>
  );
}

function JobSkeleton() {
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-24" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-5 w-16" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-2 w-32" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-20" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-24" /></td>
    </tr>
  );
}
