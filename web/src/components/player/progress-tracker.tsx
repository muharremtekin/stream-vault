'use client';

import { useEffect } from 'react';
import type { RefObject } from 'react';

import { usePlayer } from '@/lib/hooks/use-player';

interface ProgressTrackerProps {
  videoRef: RefObject<HTMLVideoElement | null>;
  contentId: string;
}

// Minimum seconds of progress before seeking on resume (avoids seeking for near-start positions)
const MIN_RESUME_POSITION = 5;

export function ProgressTracker({ videoRef, contentId }: ProgressTrackerProps) {
  const { initialProgress, saveCurrentProgress } = usePlayer(contentId);

  // Seek to saved position when initial progress data arrives and video is ready
  useEffect(() => {
    if (!initialProgress || !videoRef.current) return;
    if (initialProgress.position_seconds <= MIN_RESUME_POSITION) return;

    videoRef.current.currentTime = initialProgress.position_seconds;
  }, [initialProgress, videoRef]);

  // Save progress when the user closes or navigates away from the page
  useEffect(() => {
    const handleBeforeUnload = () => {
      saveCurrentProgress();
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [saveCurrentProgress]);

  // Save progress when the tab becomes hidden (user switches tabs or minimises)
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'hidden') {
        saveCurrentProgress();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [saveCurrentProgress]);

  return null;
}
