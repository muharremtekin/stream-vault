'use client';

import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useGenre, useGenreContent } from '@/lib/hooks/use-content';
import { cn } from '@/lib/utils/cn';

import type { ContentSummary } from '@/lib/types/catalog';
import type { ContentCardItem } from '@/lib/types/browse';

interface GenreContentProps {
  slug: string;
}

const SKELETON_COUNT = 20;

function toCardItem(item: ContentSummary): ContentCardItem {
  return {
    contentId: item.id,
    title: item.title,
    thumbnailUrl: item.thumbnailUrl,
    releaseYear: item.releaseYear,
    averageRating: item.averageRating ?? undefined,
    genres: item.genres,
    contentType: item.contentType,
  };
}

export function GenreContent({ slug }: GenreContentProps) {
  const t = useTranslations('genre');

  const { data: genre } = useGenre(slug);
  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useGenreContent(slug);

  const allItems = data?.pages.flatMap((page) => page.items) ?? [];
  const totalCount = data?.pages[0]?.totalCount ?? 0;

  if (isLoading) {
    return (
      <div className="px-4 py-8 sm:px-8 lg:px-12">
        <Skeleton variant="text" className="mb-2 h-9 w-48" />
        <Skeleton variant="text" className="mb-8 h-5 w-24" />
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
            // eslint-disable-next-line react/no-array-index-key
            <Skeleton key={`genre-card-${i}`} variant="custom" className="aspect-[2/3] w-full rounded-md" />
          ))}
        </div>
      </div>
    );
  }

  if (!isLoading && allItems.length === 0) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 px-4">
        <p className="text-lg text-muted-foreground">{t('noContent')}</p>
      </div>
    );
  }

  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      <h1 className="mb-1 text-3xl font-bold text-foreground">
        {genre?.name ?? slug}
      </h1>
      {totalCount > 0 && (
        <p className="mb-8 text-sm text-muted-foreground">
          {t('results', { count: totalCount })}
        </p>
      )}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {allItems.map((item) => (
          <ContentCard key={item.id} item={toCardItem(item)} />
        ))}
      </div>

      {hasNextPage && (
        <div className="mt-10 flex justify-center">
          <Button
            variant="secondary"
            onClick={() => void fetchNextPage()}
            className={cn(isFetchingNextPage && 'opacity-70')}
          >
            {isFetchingNextPage ? t('loadingMore') : t('loadMore')}
          </Button>
        </div>
      )}

    </div>
  );
}
