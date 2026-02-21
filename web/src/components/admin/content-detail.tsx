'use client';

import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { useTranslations } from 'next-intl';
import { ArrowLeft, Film, Tv } from 'lucide-react';

import { useAdminMovie, useAdminSeriesDetail } from '@/lib/hooks/use-admin';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { VideoUploader } from '@/components/admin/video-uploader';
import { cn } from '@/lib/utils/cn';

interface ContentDetailProps {
  contentId: string;
}

export function ContentDetail({ contentId }: ContentDetailProps) {
  const t = useTranslations('admin.contentDetail');
  const router = useRouter();

  const movieQuery = useAdminMovie(contentId);
  const movieResolved = movieQuery.isFetched;
  const movieNotFound = movieQuery.isError || (movieResolved && !movieQuery.data);
  const seriesQuery = useAdminSeriesDetail(contentId, { enabled: movieNotFound });

  const isLoading = movieQuery.isLoading || (movieNotFound && seriesQuery.isLoading);
  const movie = movieQuery.data;
  const series = seriesQuery.data;
  const content = movie ?? series;
  const isMovie = !!movie;

  if (isLoading) return <DetailSkeleton />;
  if (!content) {
    return (
      <div className="py-12 text-center">
        <p className="text-muted-foreground">{t('notFound')}</p>
        <Button variant="ghost" onClick={() => router.push('/admin/content')} className="mt-4">
          <ArrowLeft className="mr-2 h-4 w-4" /> {t('backToList')}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={() => router.push('/admin/content')} aria-label={t('backToList')}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h1 className="text-2xl font-bold text-foreground">{content.title}</h1>
        <Badge variant={isMovie ? 'info' : 'default'} size="sm">
          {isMovie ? <Film className="mr-1 h-3 w-3" /> : <Tv className="mr-1 h-3 w-3" />}
          {isMovie ? t('movie') : t('series')}
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <InfoCard label={t('description')} value={content.description} />
          <div className="grid gap-4 sm:grid-cols-3">
            <InfoCard label={t('releaseYear')} value={String(content.releaseYear)} />
            <InfoCard label={t('maturityRating')} value={content.maturityRating} />
            <InfoCard label={t('status')} value={content.status} />
          </div>
          {isMovie && movie && (
            <div className="grid gap-4 sm:grid-cols-2">
              <InfoCard label={t('director')} value={movie.director} />
              <InfoCard label={t('duration')} value={t('durationValue', { minutes: movie.durationMinutes })} />
            </div>
          )}
          {!isMovie && series && (
            <div className="grid gap-4 sm:grid-cols-3">
              <InfoCard label={t('creator')} value={series.creator} />
              <InfoCard label={t('totalSeasons')} value={String(series.totalSeasons)} />
              <InfoCard label={t('totalEpisodes')} value={String(series.totalEpisodes)} />
            </div>
          )}
          <div className="flex flex-wrap gap-1.5">
            {content.genres.map((g) => <Badge key={g} size="sm">{g}</Badge>)}
          </div>
          {content.cast.length > 0 && (
            <div className="rounded-lg border border-border p-4">
              <h3 className="mb-2 text-sm font-medium text-muted-foreground">{t('cast')}</h3>
              <div className="grid gap-1 sm:grid-cols-2">
                {content.cast.map((c) => (
                  <p key={`${c.name}-${c.role}`} className="text-sm text-foreground">
                    <span className="font-medium">{c.name}</span>
                    <span className="text-muted-foreground"> — {c.role}</span>
                  </p>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="space-y-4">
          {content.thumbnailUrl && (
            <div className="relative aspect-[2/3] overflow-hidden rounded-lg border border-border">
              <Image src={content.thumbnailUrl} alt={content.title} fill className="object-cover" sizes="(max-width: 1024px) 100vw, 33vw" unoptimized />
            </div>
          )}
        </div>
      </div>

      <div className="rounded-lg border border-border p-6">
        <h2 className="mb-4 text-lg font-semibold text-foreground">{t('videoUpload')}</h2>
        <VideoUploader contentId={contentId} />
      </div>
    </div>
  );
}

interface InfoCardProps {
  label: string;
  value: string;
}

function InfoCard({ label, value }: InfoCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={cn('mt-1 text-sm text-foreground', value.length > 100 && 'line-clamp-4')}>
        {value}
      </p>
    </div>
  );
}

function DetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Skeleton variant="custom" className="h-8 w-8 rounded-md" />
        <Skeleton variant="text" className="h-7 w-64" />
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <Skeleton variant="custom" className="h-24 w-full rounded-lg" />
          <div className="grid gap-4 sm:grid-cols-3">
            <Skeleton variant="custom" className="h-16 rounded-lg" />
            <Skeleton variant="custom" className="h-16 rounded-lg" />
            <Skeleton variant="custom" className="h-16 rounded-lg" />
          </div>
        </div>
        <Skeleton variant="card" className="h-64 w-full" />
      </div>
    </div>
  );
}
