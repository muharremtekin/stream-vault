'use client';

import { useQuery } from '@tanstack/react-query';
import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import ContentRow from '@/components/browse/content-row';
import { Skeleton } from '@/components/ui/skeleton';
import { getMovie } from '@/lib/api/catalog';
import { useContinueWatching } from '@/lib/hooks/use-browse';
import type { ContentCardItem } from '@/lib/types/browse';
import type { WatchProgress } from '@/lib/types/streaming';

interface ContinueWatchingCardProps {
  progress: WatchProgress;
}

function ContinueWatchingCard({ progress }: ContinueWatchingCardProps) {
  const { data: movie, isLoading } = useQuery({
    queryKey: ['catalog', 'movie', progress.content_id],
    queryFn: () => getMovie(progress.content_id),
    staleTime: 10 * 60_000,
  });

  if (isLoading) {
    return (
      <div className="shrink-0 w-32 sm:w-36 md:w-44">
        <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
      </div>
    );
  }

  if (!movie) return null;

  const item: ContentCardItem = {
    contentId: progress.content_id,
    title: movie.title,
    thumbnailUrl: movie.thumbnailUrl,
    releaseYear: movie.releaseYear,
    averageRating: movie.averageRating,
    genres: movie.genres,
    contentType: 'Movie',
    durationMinutes: movie.durationMinutes,
    percentage: progress.percentage,
  };

  return <ContentCard item={item} />;
}

export default function ContinueWatchingRow() {
  const t = useTranslations('browse');
  const { data, isLoading } = useContinueWatching();

  const items = data?.items ?? [];

  if (!isLoading && items.length === 0) return null;

  return (
    <ContentRow title={t('continueWatching')} isLoading={isLoading}>
      {items.map((progress) => (
        <ContinueWatchingCard key={progress.content_id} progress={progress} />
      ))}
    </ContentRow>
  );
}
