'use client';

import { useInfiniteQuery, useQuery } from '@tanstack/react-query';

import { getGenreContent, getGenres, getMovie, getSeriesById } from '@/lib/api/catalog';
import { getSimilar } from '@/lib/api/recommendation';

import type { PaginatedResponse } from '@/lib/types/common';
import type { ContentSummary, Genre, Movie, Series } from '@/lib/types/catalog';

const STALE_5MIN = 5 * 60_000;

export function useMovie(id: string) {
  return useQuery({
    queryKey: ['catalog', 'movie', id],
    queryFn: () => getMovie(id),
    staleTime: STALE_5MIN,
    enabled: !!id,
  });
}

export function useSeries(id: string) {
  return useQuery({
    queryKey: ['catalog', 'series', id],
    queryFn: () => getSeriesById(id),
    staleTime: STALE_5MIN,
    enabled: !!id,
  });
}

/** Single hook that fetches Movie OR Series depending on contentType. */
export function useContent(id: string, contentType: 'movie' | 'series') {
  return useQuery<Movie | Series>({
    queryKey: ['catalog', contentType, id],
    queryFn: () =>
      contentType === 'movie' ? getMovie(id) : getSeriesById(id),
    staleTime: STALE_5MIN,
    enabled: !!id,
  });
}

export function useSimilarContent(id: string, limit = 12) {
  return useQuery({
    queryKey: ['recommendations', 'similar', id, limit],
    queryFn: () => getSimilar(id, limit),
    staleTime: STALE_5MIN,
    enabled: !!id,
  });
}

/** Fetches all genres and selects the one matching the given slug. */
export function useGenre(slug: string) {
  return useQuery<Genre[], Error, Genre | undefined>({
    queryKey: ['catalog', 'genres'],
    queryFn: getGenres,
    select: (genres) => genres.find((g) => g.slug === slug),
    staleTime: 10 * 60_000,
    enabled: !!slug,
  });
}

/** Infinite query for paginated content within a genre. */
export function useGenreContent(slug: string, pageSize = 20) {
  return useInfiniteQuery<PaginatedResponse<ContentSummary>>({
    queryKey: ['catalog', 'genre', slug, { pageSize }],
    queryFn: ({ pageParam }) =>
      getGenreContent(slug, { page: pageParam as number, pageSize }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.hasNextPage ? lastPage.page + 1 : undefined,
    staleTime: STALE_5MIN,
    enabled: !!slug,
  });
}
