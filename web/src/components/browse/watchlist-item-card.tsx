'use client';

import { useQuery } from '@tanstack/react-query';

import ContentCard from '@/components/browse/content-card';
import { Skeleton } from '@/components/ui/skeleton';
import { getMovie, getSeriesById } from '@/lib/api/catalog';
import type { WatchlistItem } from '@/lib/api/watchlist';
import type { ContentCardItem } from '@/lib/types/browse';
import type { Movie, Series } from '@/lib/types/catalog';

interface WatchlistItemCardProps {
  item: WatchlistItem;
}

function movieToCard(movie: Movie): ContentCardItem {
  return {
    contentId: movie.id,
    title: movie.title,
    thumbnailUrl: movie.thumbnailUrl,
    releaseYear: movie.releaseYear,
    averageRating: movie.averageRating,
    genres: movie.genres,
    contentType: 'Movie',
    durationMinutes: movie.durationMinutes,
  };
}

function seriesToCard(series: Series): ContentCardItem {
  return {
    contentId: series.id,
    title: series.title,
    thumbnailUrl: series.thumbnailUrl,
    releaseYear: series.releaseYear,
    averageRating: series.averageRating,
    genres: series.genres,
    contentType: 'Series',
    totalSeasons: series.totalSeasons,
  };
}

export default function WatchlistItemCard({ item }: WatchlistItemCardProps) {
  const isMovie = item.contentType !== 'Series';

  const { data, isLoading, isError } = useQuery<Movie | Series>({
    queryKey: isMovie
      ? ['catalog', 'movie', item.contentId]
      : ['catalog', 'series', item.contentId],
    queryFn: () =>
      isMovie ? getMovie(item.contentId) : getSeriesById(item.contentId),
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) {
    return (
      <div className="shrink-0 w-32 sm:w-36 md:w-44">
        <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
      </div>
    );
  }

  if (isError || !data) {
    return null;
  }

  const cardItem = isMovie
    ? movieToCard(data as Movie)
    : seriesToCard(data as Series);

  return <ContentCard item={cardItem} />;
}
