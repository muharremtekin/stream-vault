'use client';

import { useTranslations } from 'next-intl';

import { useAdminEncodingJobs } from '@/lib/hooks/use-admin';
import { Badge } from '@/components/ui/badge';
import { ProgressBar } from '@/components/ui/progress-bar';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDateTime } from '@/lib/utils/format';

import type { EncodingJob } from '@/lib/types/encoding';
import { ENCODING_STATUS_BADGE, ENCODING_PROGRESS_VARIANT } from '@/lib/types/encoding';

const STATUS_I18N_MAP = {
  pending: 'pending',
  processing: 'processing',
  completed: 'completed',
  failed: 'failed',
} as const;

function JobSkeleton() {
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-24" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-5 w-16" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-2 w-full" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-20" /></td>
    </tr>
  );
}

export function RecentJobs() {
  const t = useTranslations('admin.dashboard');
  const { data, isLoading } = useAdminEncodingJobs({ limit: 5 });

  const jobs = data?.jobs ?? [];

  return (
    <div>
      <h2 className="mb-4 text-lg font-semibold text-foreground">{t('recentJobs')}</h2>
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('contentId')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('status')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('progress')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('createdAt')}</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <>
                <JobSkeleton />
                <JobSkeleton />
                <JobSkeleton />
                <JobSkeleton />
                <JobSkeleton />
              </>
            ) : jobs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                  {t('noRecentJobs')}
                </td>
              </tr>
            ) : (
              jobs.map((job: EncodingJob) => {
                const statusKey = job.status.toLowerCase();
                return (
                  <tr key={job.jobId} className="border-b border-border last:border-b-0 hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-xs text-foreground">
                      {job.contentId.length > 8 ? `${job.contentId.slice(0, 8)}...` : job.contentId}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={ENCODING_STATUS_BADGE[statusKey] ?? 'default'} size="sm">
                        {statusKey in STATUS_I18N_MAP
                          ? t(STATUS_I18N_MAP[statusKey as keyof typeof STATUS_I18N_MAP])
                          : job.status}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <ProgressBar
                        value={job.progressPercentage}
                        variant={ENCODING_PROGRESS_VARIANT[statusKey] ?? 'primary'}
                        size="sm"
                        showLabel
                      />
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {formatDateTime(job.createdAt)}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
