'use client';

import { useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { useAdminMovies, useAdminSeries } from '@/lib/hooks/use-admin';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { formatShortDate } from '@/lib/utils/format';

import type { Movie, Series } from '@/lib/types/catalog';

interface ContentRow {
  id: string;
  title: string;
  type: 'movie' | 'series';
  status: string;
  createdAt: string;
}

function mapMoviesToRows(movies: Movie[]): ContentRow[] {
  return movies.map((m) => ({
    id: m.id,
    title: m.title,
    type: 'movie' as const,
    status: m.status,
    createdAt: m.createdAt,
  }));
}

function mapSeriesToRows(series: Series[]): ContentRow[] {
  return series.map((s) => ({
    id: s.id,
    title: s.title,
    type: 'series' as const,
    status: s.status,
    createdAt: s.createdAt,
  }));
}

const TYPE_BADGE_MAP: Record<string, 'info' | 'default'> = {
  movie: 'info',
  series: 'default',
};

const STATUS_BADGE_MAP: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
  Published: 'success',
  Draft: 'default',
  Active: 'success',
  Inactive: 'warning',
};

interface ContentTableProps {
  search: string;
}

function RowSkeleton() {
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-40" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-5 w-14" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-5 w-16" /></td>
      <td className="px-4 py-3"><Skeleton variant="text" className="h-4 w-20" /></td>
    </tr>
  );
}

export function ContentTable({ search }: ContentTableProps) {
  const t = useTranslations('admin.content');
  const router = useRouter();

  const { data: moviesData, isLoading: isLoadingMovies } = useAdminMovies({ pageSize: 100 });
  const { data: seriesData, isLoading: isLoadingSeries } = useAdminSeries({ pageSize: 100 });

  const isLoading = isLoadingMovies || isLoadingSeries;

  const rows = useMemo(() => {
    const movieRows = mapMoviesToRows(moviesData?.items ?? []);
    const seriesRows = mapSeriesToRows(seriesData?.items ?? []);
    const allRows = [...movieRows, ...seriesRows];

    allRows.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

    if (!search.trim()) return allRows;
    const query = search.toLowerCase().trim();
    return allRows.filter((row) => row.title.toLowerCase().includes(query));
  }, [moviesData, seriesData, search]);

  const handleRowKeyDown = (e: React.KeyboardEvent, id: string) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      router.push(`/admin/content/${id}`);
    }
  };

  return (
    <div>
      {!isLoading && (
        <p className="mb-2 text-sm text-muted-foreground">
          {t('showingResults', { count: rows.length })}
        </p>
      )}
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('title')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('type')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('status')}</th>
              <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('date')}</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <>
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
                <RowSkeleton />
              </>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center">
                  <p className="text-muted-foreground">{t('noContent')}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{t('noContentDescription')}</p>
                </td>
              </tr>
            ) : (
              rows.map((row) => (
                <tr
                  key={row.id}
                  tabIndex={0}
                  onClick={() => router.push(`/admin/content/${row.id}`)}
                  onKeyDown={(e) => handleRowKeyDown(e, row.id)}
                  className="cursor-pointer border-b border-border transition-colors last:border-b-0 hover:bg-muted/30 focus-visible:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                >
                  <td className="px-4 py-3 font-medium text-foreground">{row.title}</td>
                  <td className="px-4 py-3">
                    <Badge variant={TYPE_BADGE_MAP[row.type] ?? 'default'} size="sm">
                      {t(row.type)}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={STATUS_BADGE_MAP[row.status] ?? 'default'} size="sm">
                      {row.status}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">
                    {formatShortDate(row.createdAt)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
