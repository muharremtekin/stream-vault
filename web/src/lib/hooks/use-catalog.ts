'use client';

import { useInfiniteQuery, useQuery } from '@tanstack/react-query';

import { getGenres, getMovies, getSeries } from '@/lib/api/catalog';

import type { CatalogParams, Series } from '@/lib/types/catalog';
import type { PaginatedResponse } from '@/lib/types/common';

const STALE_5MIN = 5 * 60_000;
const STALE_10MIN = 10 * 60_000;
const DEFAULT_PAGE_SIZE = 20;

export function useMovies(params: CatalogParams) {
  return useQuery({
    queryKey: ['catalog', 'movies', params],
    queryFn: () => getMovies(params),
    staleTime: STALE_5MIN,
    placeholderData: (previousData) => previousData,
  });
}

export function useMovieGenres() {
  return useQuery({
    queryKey: ['catalog', 'genres'],
    queryFn: getGenres,
    staleTime: STALE_10MIN,
  });
}

export function useSeriesCatalog(params?: Omit<CatalogParams, 'page'>) {
  const pageSize = params?.pageSize ?? DEFAULT_PAGE_SIZE;

  return useInfiniteQuery<PaginatedResponse<Series>>({
    queryKey: ['catalog', 'series', { ...params, pageSize }],
    queryFn: ({ pageParam }) =>
      getSeries({ ...params, page: pageParam as number, pageSize }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.hasNextPage ? lastPage.page + 1 : undefined,
    staleTime: STALE_5MIN,
  });
}
