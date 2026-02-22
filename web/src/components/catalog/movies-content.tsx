'use client';

import { useCallback } from 'react';

import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { MoviesFilterBar } from '@/components/catalog/movies-filter-bar';
import { MoviesGrid } from '@/components/catalog/movies-grid';
import { SearchPagination } from '@/components/search/search-pagination';
import { useMovies, useMovieGenres } from '@/lib/hooks/use-catalog';
import { buildMovieParams, movieToCardItem } from '@/lib/utils/catalog-helpers';

import type { CatalogParams } from '@/lib/types/catalog';

export function MoviesContent() {
  const t = useTranslations('pages.movies');
  const router = useRouter();
  const pathname = usePathname();
  const urlSearchParams = useSearchParams();

  const params = buildMovieParams(urlSearchParams);
  const { data, isLoading, isFetching, isError } = useMovies(params);
  const { data: genres } = useMovieGenres();

  const hasActiveFilters = !!params.genre || (!!params.sort && params.sort !== 'newest');

  const handleFilterChange = useCallback(
    (updates: Partial<CatalogParams>) => {
      const newParams = new URLSearchParams(urlSearchParams.toString());

      for (const [key, value] of Object.entries(updates)) {
        if (value === undefined || value === '') {
          newParams.delete(key);
        } else {
          newParams.set(key, String(value));
        }
      }

      // Reset to page 1 when filters change (not when page itself changes)
      if (!('page' in updates)) {
        newParams.delete('page');
      }

      // Remove page=1 from URL for cleanliness
      if (newParams.get('page') === '1') {
        newParams.delete('page');
      }

      const search = newParams.toString();
      router.replace(search ? `${pathname}?${search}` : pathname);
    },
    [urlSearchParams, pathname, router]
  );

  const handlePageChange = useCallback(
    (page: number) => {
      handleFilterChange({ page } as Partial<CatalogParams>);
    },
    [handleFilterChange]
  );

  const handleClearAll = useCallback(() => {
    router.replace(pathname);
  }, [pathname, router]);

  if (isError) {
    return (
      <div className="flex min-h-[40vh] flex-col items-center justify-center gap-4 text-center">
        <p className="text-muted-foreground">{t('noResultsHint')}</p>
        <button
          type="button"
          onClick={handleClearAll}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          {t('clearFilters')}
        </button>
      </div>
    );
  }

  const items = data?.items.map(movieToCardItem) ?? [];
  const showSkeleton = isLoading;

  return (
    <>
      <MoviesFilterBar
        genres={genres ?? []}
        activeGenre={params.genre}
        activeSort={params.sort}
        onFilterChange={handleFilterChange}
        isLoading={isFetching}
      />

      {data && !showSkeleton && data.totalCount > 0 && (
        <p className="mb-6 text-sm text-muted-foreground">
          {t('resultsCount', { count: data.totalCount })}
        </p>
      )}

      <MoviesGrid
        items={items}
        isLoading={showSkeleton}
        onClearFilters={handleClearAll}
        hasActiveFilters={hasActiveFilters}
      />

      {data && data.totalPages > 1 && (
        <div className="mt-10">
          <SearchPagination
            currentPage={data.page}
            totalPages={data.totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      )}
    </>
  );
}
