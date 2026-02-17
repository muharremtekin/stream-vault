'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import {
  fetchWatchlist,
  addToWatchlist,
  removeFromWatchlist,
} from '@/lib/api/watchlist';
import { useAuthStore } from '@/lib/stores/auth-store';
import type { ContentType } from '@/lib/types/common';
import type { WatchlistItem } from '@/lib/api/watchlist';

export function useWatchlist() {
  const queryClient = useQueryClient();
  const activeProfile = useAuthStore((s) => s.activeProfile);

  const watchlistQuery = useQuery({
    queryKey: ['watchlist', 'list', activeProfile?.id],
    queryFn: () => fetchWatchlist(activeProfile?.id ?? ''),
    enabled: !!activeProfile?.id,
    staleTime: 60_000,
  });

  const addMutation = useMutation({
    mutationFn: ({
      contentId,
      contentType,
    }: {
      contentId: string;
      contentType: ContentType;
    }) => addToWatchlist(activeProfile?.id ?? '', contentId, contentType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['watchlist', 'list'] });
    },
  });

  const removeMutation = useMutation({
    mutationFn: (id: string) => removeFromWatchlist(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['watchlist', 'list'] });
    },
  });

  const isInWatchlist = (contentId: string): boolean =>
    watchlistQuery.data?.some((item) => item.contentId === contentId) ?? false;

  const getWatchlistItemId = (contentId: string): string | undefined =>
    watchlistQuery.data?.find((item) => item.contentId === contentId)?.id;

  return {
    items: watchlistQuery.data ?? [],
    isLoading: watchlistQuery.isLoading,
    add: addMutation.mutateAsync,
    remove: removeMutation.mutateAsync,
    isAdding: addMutation.isPending,
    isRemoving: removeMutation.isPending,
    isInWatchlist,
    getWatchlistItemId,
  };
}
