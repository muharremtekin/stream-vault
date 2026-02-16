'use client';

import { useEffect, useRef, useCallback } from 'react';

import { useQuery } from '@tanstack/react-query';

import { getProgress, saveProgress } from '@/lib/api/streaming';
import { usePlayerStore } from '@/lib/stores/player-store';
import { PROGRESS_SAVE_INTERVAL } from '@/lib/utils/constants';

export function usePlayer(contentId: string | null) {
  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const duration = usePlayerStore((s) => s.duration);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const lastSavedRef = useRef<number>(0);

  const progressQuery = useQuery({
    queryKey: ['streaming', 'progress', contentId],
    queryFn: () => getProgress(contentId ?? ''),
    enabled: !!contentId,
    staleTime: 30_000,
  });

  const saveCurrentProgress = useCallback(async () => {
    if (!contentId || duration <= 0) return;

    const currentTime = usePlayerStore.getState().currentTime;
    const currentDuration = usePlayerStore.getState().duration;

    // Skip if position hasn't changed
    if (Math.abs(currentTime - lastSavedRef.current) < 1) return;

    try {
      await saveProgress(contentId, {
        position_seconds: Math.floor(currentTime),
        duration_seconds: Math.floor(currentDuration),
      });
      lastSavedRef.current = currentTime;
    } catch {
      // Silently ignore save errors to avoid disrupting playback
    }
  }, [contentId, duration]);

  // Auto-save progress at regular intervals while playing
  useEffect(() => {
    if (isPlaying && contentId) {
      intervalRef.current = setInterval(
        saveCurrentProgress,
        PROGRESS_SAVE_INTERVAL
      );
    }

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [isPlaying, contentId, saveCurrentProgress]);

  // Save progress on unmount
  useEffect(() => {
    return () => {
      saveCurrentProgress();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    initialProgress: progressQuery.data,
    isLoadingProgress: progressQuery.isLoading,
    saveCurrentProgress,
  };
}
