import { apiClient } from '@/lib/api/client';

import type { ContentType } from '@/lib/types/common';

export interface WatchlistItem {
  id: string;
  contentId: string;
  contentType: string;
  addedAt: string;
  note: string | null;
}

export async function fetchWatchlist(profileId: string): Promise<WatchlistItem[]> {
  const response = await apiClient.get<WatchlistItem[]>(
    `/api/watchlist/profile/${profileId}`
  );
  return response.data;
}

export async function addToWatchlist(
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

export async function removeFromWatchlist(id: string): Promise<void> {
  await apiClient.delete(`/api/watchlist/${id}`);
}
