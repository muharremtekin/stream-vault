'use client';

import axios from 'axios';
import { useQuery } from '@tanstack/react-query';
import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import ContentRow from '@/components/browse/content-row';
import { Skeleton } from '@/components/ui/skeleton';
import { getMovie, getSeriesById } from '@/lib/api/catalog';
import { useContinueWatching } from '@/lib/hooks/use-browse';
import type { ContentCardItem } from '@/lib/types/browse';
import type { Movie, Series } from '@/lib/types/catalog';
import type { WatchProgress } from '@/lib/types/streaming';
import {
  buildEpisodeWatchHref,
  resolveCatalogReference,
} from '@/lib/utils/episode-content-id';

interface ContinueWatchingCardProps {
  progress: WatchProgress;
}

function ContinueWatchingCard({ progress }: ContinueWatchingCardProps) {
  const catalogReference = resolveCatalogReference(progress.content_id);
  const { data: content, isLoading } = useQuery<Movie | Series>({
    queryKey: [
      'catalog',
      catalogReference.contentType,
      catalogReference.contentId,
    ],
    queryFn: () =>
      catalogReference.contentType === 'series'
        ? getSeriesById(catalogReference.contentId)
        : getMovie(catalogReference.contentId),
    staleTime: 10 * 60_000,
    retry: (failureCount, error) => {
      const status = axios.isAxiosError(error) ? error.response?.status : undefined;
      return (status === undefined || status >= 500) && failureCount < 1;
    },
  });

  if (isLoading) {
    return (
      <div className="shrink-0 w-32 sm:w-36 md:w-44">
        <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
      </div>
    );
  }

  if (!content) return null;

  if (catalogReference.contentType === 'series') {
    if (!('seasons' in content)) return null;

    const episodeReference = catalogReference.episode;
    const season = content.seasons.find(
      (item) => item.seasonNumber === episodeReference.seasonNumber,
    );
    const episode = season?.episodes.find(
      (item) => item.episodeNumber === episodeReference.episodeNumber,
    );
    const title = episode?.title ?? content.title;

    const item: ContentCardItem = {
      contentId: content.id,
      playbackHref: buildEpisodeWatchHref(progress.content_id, title),
      title,
      thumbnailUrl: episode?.thumbnailUrl || content.thumbnailUrl || undefined,
      releaseYear: season?.releaseYear ?? content.releaseYear,
      averageRating: content.averageRating,
      genres: content.genres,
      contentType: 'Series',
      durationMinutes: episode?.durationMinutes,
      totalSeasons: content.totalSeasons,
      percentage: progress.percentage,
    };

    return <ContentCard item={item} />;
  }

  if (!('durationMinutes' in content)) return null;

  const item: ContentCardItem = {
    contentId: content.id,
    title: content.title,
    thumbnailUrl: content.thumbnailUrl,
    releaseYear: content.releaseYear,
    averageRating: content.averageRating,
    genres: content.genres,
    contentType: 'Movie',
    durationMinutes: content.durationMinutes,
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
