'use client';

import { useCallback } from 'react';

import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { Search } from 'lucide-react';
import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import { SearchFilters } from '@/components/search/search-filters';
import { SearchPagination } from '@/components/search/search-pagination';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useSearch } from '@/lib/hooks/use-search';
import { useTrending } from '@/lib/hooks/use-browse';
import {
  buildSearchParams,
  searchHitToCardItem,
  trendingItemToCardItem,
} from '@/lib/utils/search-helpers';
import type { SearchParams } from '@/lib/types/search';

const SKELETON_COUNT = 24;

export function SearchResults() {
  const t = useTranslations('search');
  const router = useRouter();
  const pathname = usePathname();
  const urlSearchParams = useSearchParams();

  const params = buildSearchParams(urlSearchParams);
  const hasQuery = !!params.q && params.q.length > 0;

  const { data, isLoading, isError } = useSearch(params);
  const { data: trendingData, isLoading: isTrendingLoading } = useTrending('week', 20);

  const handleFilterChange = useCallback(
    (updates: Partial<SearchParams>) => {
      const newParams = new URLSearchParams(urlSearchParams.toString());

      for (const [key, value] of Object.entries(updates)) {
        if (value === undefined || value === '') {
          newParams.delete(key);
        } else {
          newParams.set(key, String(value));
        }
      }

      if (!('page' in updates)) {
        newParams.delete('page');
      }

      router.replace(`${pathname}?${newParams.toString()}`);
    },
    [urlSearchParams, pathname, router]
  );

  const handlePageChange = useCallback(
    (page: number) => {
      handleFilterChange({ page } as Partial<SearchParams>);
    },
    [handleFilterChange]
  );

  const handleClearAll = () => {
    router.replace(pathname);
  };

  if (!hasQuery) {
    return <TrendingFallback data={trendingData} isLoading={isTrendingLoading} />;
  }

  return (
    <div className="flex flex-col gap-6 md:flex-row">
      {/* Filters sidebar */}
      <div className="w-full shrink-0 md:w-56 lg:w-64">
        <SearchFilters
          facets={data?.facets}
          currentParams={params}
          onFilterChange={handleFilterChange}
        />
      </div>

      {/* Results area */}
      <div className="min-w-0 flex-1">
        {isLoading && <ResultsSkeleton />}

        {!isLoading && isError && (
          <div className="flex min-h-[40vh] flex-col items-center justify-center gap-4 text-center">
            <p className="text-muted-foreground">{t('noResultsHint')}</p>
            <Button variant="secondary" onClick={handleClearAll}>
              {t('clearFilters')}
            </Button>
          </div>
        )}

        {!isLoading && !isError && data && data.items.length === 0 && (
          <EmptyResults query={params.q ?? ''} onClear={handleClearAll} />
        )}

        {!isLoading && !isError && data && data.items.length > 0 && (
          <>
            <p className="mb-6 text-sm text-muted-foreground">
              {t('resultsCount', { count: data.totalCount })}
            </p>

            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
              {data.items.map((hit) => (
                <ContentCard key={hit.contentId} item={searchHitToCardItem(hit)} />
              ))}
            </div>

            <div className="mt-10">
              <SearchPagination
                currentPage={data.page}
                totalPages={data.totalPages}
                onPageChange={handlePageChange}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function EmptyResults({ query, onClear }: { query: string; onClear: () => void }) {
  const t = useTranslations('search');

  return (
    <div className="flex min-h-[40vh] flex-col items-center justify-center gap-4 text-center">
      <Search className="h-16 w-16 text-muted-foreground/40" />
      <p className="text-lg font-semibold text-foreground">
        {t('noResults', { query })}
      </p>
      <p className="text-sm text-muted-foreground">{t('noResultsHint')}</p>
      <Button variant="secondary" onClick={onClear}>
        {t('clearFilters')}
      </Button>
    </div>
  );
}

function ResultsSkeleton() {
  return (
    <>
      <Skeleton variant="text" className="mb-6 h-5 w-24" />
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
          <Skeleton
            key={`search-skeleton-${String(i)}`}
            variant="custom"
            className="aspect-[2/3] w-full rounded-md"
          />
        ))}
      </div>
    </>
  );
}

function TrendingFallback({
  data,
  isLoading,
}: {
  data: { items: Array<{ contentId: string; title: string; thumbnailUrl: string; contentType: string; releaseYear: number; averageRating: number; rank: number; rankChange: number; viewCount: number }> } | undefined;
  isLoading: boolean;
}) {
  const t = useTranslations('search');

  return (
    <div>
      <h1 className="mb-2 text-2xl font-bold text-foreground">{t('trending')}</h1>
      <p className="mb-8 text-sm text-muted-foreground">{t('trendingDescription')}</p>

      {isLoading && <ResultsSkeleton />}

      {!isLoading && data && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {data.items.map((item) => (
            <ContentCard key={item.contentId} item={trendingItemToCardItem(item)} showRank />
          ))}
        </div>
      )}
    </div>
  );
}
