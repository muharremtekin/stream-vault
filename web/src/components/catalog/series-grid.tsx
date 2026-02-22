'use client';

import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useSeriesCatalog } from '@/lib/hooks/use-catalog';
import { cn } from '@/lib/utils/cn';

import type { Series } from '@/lib/types/catalog';
import type { ContentCardItem } from '@/lib/types/browse';

const SKELETON_COUNT = 20;

function seriesToCardItem(item: Series): ContentCardItem {
  return {
    contentId: item.id,
    title: item.title,
    thumbnailUrl: item.thumbnailUrl,
    releaseYear: item.releaseYear,
    averageRating: item.averageRating,
    genres: item.genres,
    contentType: 'Series',
    totalSeasons: item.totalSeasons,
  };
}

export function SeriesGrid() {
  const t = useTranslations('pages.series');
  const tGenre = useTranslations('genre');

  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useSeriesCatalog();

  const allItems = data?.pages.flatMap((page) => page.items) ?? [];
  const totalCount = data?.pages[0]?.totalCount ?? 0;

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          // eslint-disable-next-line react/no-array-index-key
          <Skeleton key={`series-skeleton-${i}`} variant="custom" className="aspect-[2/3] w-full rounded-md" />
        ))}
      </div>
    );
  }

  if (allItems.length === 0) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 px-4">
        <p className="text-lg text-muted-foreground">{t('noContent')}</p>
      </div>
    );
  }

  return (
    <>
      {totalCount > 0 && (
        <p className="mb-6 text-sm text-muted-foreground">
          {tGenre('results', { count: totalCount })}
        </p>
      )}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {allItems.map((item) => (
          <ContentCard key={item.id} item={seriesToCardItem(item)} />
        ))}
      </div>

      {hasNextPage && (
        <div className="mt-10 flex justify-center">
          <Button
            variant="secondary"
            onClick={() => void fetchNextPage()}
            className={cn(isFetchingNextPage && 'opacity-70')}
          >
            {isFetchingNextPage ? tGenre('loadingMore') : tGenre('loadMore')}
          </Button>
        </div>
      )}
    </>
  );
}
