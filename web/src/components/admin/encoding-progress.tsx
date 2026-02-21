'use client';

import { useTranslations } from 'next-intl';
import { ArrowLeft, AlertTriangle } from 'lucide-react';

import { useAdminEncodingJob } from '@/lib/hooks/use-admin';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ProgressBar } from '@/components/ui/progress-bar';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDateTime, formatFileSize } from '@/lib/utils/format';

import type { EncodingOutput } from '@/lib/types/encoding';
import { ENCODING_STATUS_BADGE } from '@/lib/types/encoding';

interface EncodingProgressProps {
  jobId: string;
  onBack: () => void;
}

function formatOptionalDate(dateStr: string | undefined): string {
  if (!dateStr) return '—';
  return formatDateTime(dateStr, { showSeconds: true });
}

export function EncodingProgress({ jobId, onBack }: EncodingProgressProps) {
  const t = useTranslations('admin.encoding');
  const { data: job, isLoading } = useAdminEncodingJob(jobId);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton variant="text" className="h-7 w-48" />
        <Skeleton variant="custom" className="h-4 w-full rounded" />
        <Skeleton variant="custom" className="h-32 w-full rounded-lg" />
      </div>
    );
  }

  if (!job) {
    return (
      <div className="py-12 text-center">
        <p className="text-muted-foreground">{t('jobNotFound')}</p>
        <Button variant="ghost" onClick={onBack} className="mt-4">
          <ArrowLeft className="mr-2 h-4 w-4" /> {t('backToList')}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={onBack} aria-label={t('backToList')}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="text-xl font-bold text-foreground">{t('jobDetail')}</h2>
        <Badge variant={ENCODING_STATUS_BADGE[job.status] ?? 'default'}>{t(job.status)}</Badge>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">{t('progress')}</span>
          <span className="font-medium text-foreground">{job.progressPercentage}%</span>
        </div>
        <ProgressBar
          value={job.progressPercentage}
          variant={job.status === 'completed' ? 'success' : job.status === 'failed' ? 'warning' : 'primary'}
          size="lg"
        />
        {job.currentStep && (
          <p className="text-xs text-muted-foreground">{t('currentStep')}: {job.currentStep}</p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <InfoField label={t('contentId')} value={job.contentId} />
        <InfoField label={t('createdAt')} value={formatOptionalDate(job.createdAt)} />
        <InfoField label={t('startedAt')} value={formatOptionalDate(job.startedAt)} />
        <InfoField label={t('completedAt')} value={formatOptionalDate(job.completedAt)} />
      </div>

      {job.status === 'failed' && job.errorMessage && (
        <div className="flex items-start gap-3 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
          <AlertTriangle className="mt-0.5 h-5 w-5 text-destructive" />
          <div>
            <p className="text-sm font-medium text-foreground">{t('errorMessage')}</p>
            <p className="mt-1 text-sm text-muted-foreground">{job.errorMessage}</p>
          </div>
        </div>
      )}

      {job.outputs && job.outputs.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-medium text-foreground">{t('outputs')}</h3>
          <div className="overflow-x-auto rounded-lg border border-border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/50">
                  <th className="px-4 py-2 text-left font-medium text-muted-foreground">{t('quality')}</th>
                  <th className="px-4 py-2 text-left font-medium text-muted-foreground">{t('resolution')}</th>
                  <th className="px-4 py-2 text-left font-medium text-muted-foreground">{t('bitrate')}</th>
                  <th className="px-4 py-2 text-left font-medium text-muted-foreground">{t('fileSize')}</th>
                  <th className="px-4 py-2 text-left font-medium text-muted-foreground">{t('segments')}</th>
                </tr>
              </thead>
              <tbody>
                {job.outputs.map((o: EncodingOutput) => (
                  <tr key={o.quality} className="border-b border-border last:border-b-0">
                    <td className="px-4 py-2 font-medium text-foreground">{o.quality}</td>
                    <td className="px-4 py-2 text-muted-foreground">{o.width}x{o.height}</td>
                    <td className="px-4 py-2 text-muted-foreground">{o.bitrateKbps} kbps</td>
                    <td className="px-4 py-2 text-muted-foreground">{formatFileSize(o.fileSizeBytes)}</td>
                    <td className="px-4 py-2 text-muted-foreground">{o.segmentCount}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

interface InfoFieldProps {
  label: string;
  value: string;
}

function InfoField({ label, value }: InfoFieldProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 truncate text-sm font-mono text-foreground" title={value}>{value}</p>
    </div>
  );
}
