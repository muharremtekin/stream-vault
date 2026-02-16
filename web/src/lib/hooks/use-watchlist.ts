'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { apiClient } from '@/lib/api/client';
import { useAuthStore } from '@/lib/stores/auth-store';
import type { ContentType } from '@/lib/types/common';

interface WatchlistItem {
  id: string;
  contentId: string;
  contentType: string;
  addedAt: string;
  note: string | null;
}

async function fetchWatchlist(profileId: string): Promise<WatchlistItem[]> {
  const response = await apiClient.get<WatchlistItem[]>(
    `/api/watchlist/profile/${profileId}`
  );
  return response.data;
}

async function addToWatchlist(
  profileId: string,
  contentId: string,
  contentType: ContentType
): Promise<WatchlistItem> {
  const response = await apiClient.post<WatchlistItem>('/api/watchlist', {
    profileId,
    contentId,
    contentType,
  });
  return response.data;
}

async function removeFromWatchlist(id: string): Promise<void> {
  await apiClient.delete(`/api/watchlist/${id}`);
}

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
