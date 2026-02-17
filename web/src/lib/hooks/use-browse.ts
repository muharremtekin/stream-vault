'use client';

import { useQuery } from '@tanstack/react-query';

import { getMovies } from '@/lib/api/catalog';
import { getHomeSections } from '@/lib/api/recommendation';
import { getTrending } from '@/lib/api/search';
import { getContinueWatching } from '@/lib/api/streaming';
import { useAuthStore } from '@/lib/stores/auth-store';

export function useFeaturedMovie() {
  return useQuery({
    queryKey: ['catalog', 'movies', 'featured'],
    queryFn: () => getMovies({ pageSize: 10 }),
    staleTime: 10 * 60_000,
    select: (data) => data.items[0] ?? null,
  });
}

export function useHomeSections() {
  return useQuery({
    queryKey: ['recommendations', 'home'],
    queryFn: getHomeSections,
    staleTime: 5 * 60_000,
  });
}

export function useContinueWatching(limit = 10) {
  const activeProfile = useAuthStore((s) => s.activeProfile);

  return useQuery({
    queryKey: ['streaming', 'continue-watching', activeProfile?.id],
    queryFn: () => getContinueWatching(limit),
    enabled: !!activeProfile?.id,
    staleTime: 60_000,
  });
}

export function useTrending(window = 'week', limit = 10) {
  return useQuery({
    queryKey: ['search', 'trending', window],
    queryFn: () => getTrending(window, limit),
    staleTime: 5 * 60_000,
  });
}
