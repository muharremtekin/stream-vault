'use client';

import { Film, Tv, Clapperboard, HardDrive } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useAdminMovies, useAdminSeries, useAdminEncodingJobs } from '@/lib/hooks/use-admin';
import { Skeleton } from '@/components/ui/skeleton';

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: number | undefined;
  isLoading: boolean;
}

function StatCard({ icon, label, value, isLoading }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-5">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
          {icon}
        </div>
        <div className="flex-1">
          <p className="text-sm text-muted-foreground">{label}</p>
          {isLoading ? (
            <Skeleton variant="text" className="mt-1 h-7 w-16" />
          ) : (
            <p className="text-2xl font-bold text-foreground">{value ?? 0}</p>
          )}
        </div>
      </div>
    </div>
  );
}

export function StatsCards() {
  const t = useTranslations('admin.dashboard');

  const { data: moviesData, isLoading: isLoadingMovies } = useAdminMovies({ pageSize: 1 });
  const { data: seriesData, isLoading: isLoadingSeries } = useAdminSeries({ pageSize: 1 });
  const { data: jobsData, isLoading: isLoadingJobs } = useAdminEncodingJobs({ limit: 1 });

  const totalContent = (moviesData?.totalCount ?? 0) + (seriesData?.totalCount ?? 0);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard
        icon={<Film className="h-5 w-5 text-primary" />}
        label={t('totalMovies')}
        value={moviesData?.totalCount}
        isLoading={isLoadingMovies}
      />
      <StatCard
        icon={<Tv className="h-5 w-5 text-primary" />}
        label={t('totalSeries')}
        value={seriesData?.totalCount}
        isLoading={isLoadingSeries}
      />
      <StatCard
        icon={<Clapperboard className="h-5 w-5 text-primary" />}
        label={t('totalContent')}
        value={totalContent}
        isLoading={isLoadingMovies || isLoadingSeries}
      />
      <StatCard
        icon={<HardDrive className="h-5 w-5 text-primary" />}
        label={t('encodingJobs')}
        value={jobsData?.totalCount}
        isLoading={isLoadingJobs}
      />
    </div>
  );
}
