'use client';

import { Film } from 'lucide-react';
import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';

import type { ContentCardItem } from '@/lib/types/browse';

const SKELETON_COUNT = 24;

interface MoviesGridProps {
  items: ContentCardItem[];
  isLoading: boolean;
  onClearFilters?: () => void;
  hasActiveFilters?: boolean;
}

export function MoviesGrid({ items, isLoading, onClearFilters, hasActiveFilters }: MoviesGridProps) {
  const t = useTranslations('pages.movies');

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          <Skeleton key={`movie-skeleton-${String(i)}`} variant="custom" className="aspect-[2/3] w-full rounded-md" />
        ))}
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="flex min-h-[40vh] flex-col items-center justify-center gap-4 text-center">
        <Film className="h-16 w-16 text-muted-foreground/40" />
        <p className="text-lg font-semibold text-foreground">
          {hasActiveFilters ? t('noResults') : t('noContent')}
        </p>
        {hasActiveFilters && (
          <>
            <p className="text-sm text-muted-foreground">{t('noResultsHint')}</p>
            {onClearFilters && (
              <Button variant="secondary" onClick={onClearFilters}>
                {t('clearFilters')}
              </Button>
            )}
          </>
        )}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {items.map((item) => (
        <ContentCard key={item.contentId} item={item} />
      ))}
    </div>
  );
}
