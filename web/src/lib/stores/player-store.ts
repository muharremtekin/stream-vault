'use client';

import { create } from 'zustand';

interface PlayerState {
  contentId: string | null;
  contentType: string | null;
  episodeId: string | null;
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  currentQuality: string | null;
  availableQualities: string[];
  volume: number;
  isMuted: boolean;
  isFullscreen: boolean;
  isBuffering: boolean;
  playbackRate: number;
  showControls: boolean;

  setContent: (
    contentId: string,
    contentType: string,
    episodeId?: string
  ) => void;
  setIsPlaying: (isPlaying: boolean) => void;
  setCurrentTime: (time: number) => void;
  setDuration: (duration: number) => void;
  setCurrentQuality: (quality: string) => void;
  setAvailableQualities: (qualities: string[]) => void;
  setVolume: (volume: number) => void;
  setIsMuted: (isMuted: boolean) => void;
  setIsFullscreen: (isFullscreen: boolean) => void;
  setIsBuffering: (isBuffering: boolean) => void;
  setPlaybackRate: (rate: number) => void;
  setShowControls: (show: boolean) => void;
  reset: () => void;
}

const initialState = {
  contentId: null,
  contentType: null,
  episodeId: null,
  isPlaying: false,
  currentTime: 0,
  duration: 0,
  currentQuality: null,
  availableQualities: [] as string[],
  volume: 1,
  isMuted: false,
  isFullscreen: false,
  isBuffering: false,
  playbackRate: 1,
  showControls: true,
};

export const usePlayerStore = create<PlayerState>()((set) => ({
  ...initialState,

  setContent: (contentId, contentType, episodeId) =>
    set({
      contentId,
      contentType,
      episodeId: episodeId ?? null,
      currentTime: 0,
      duration: 0,
      isPlaying: false,
      isBuffering: false,
    }),

  setIsPlaying: (isPlaying) => set({ isPlaying }),
  setCurrentTime: (currentTime) => set({ currentTime }),
  setDuration: (duration) => set({ duration }),
  setCurrentQuality: (currentQuality) => set({ currentQuality }),
  setAvailableQualities: (availableQualities) => set({ availableQualities }),
  setVolume: (volume) => set({ volume }),
  setIsMuted: (isMuted) => set({ isMuted }),
  setIsFullscreen: (isFullscreen) => set({ isFullscreen }),
  setIsBuffering: (isBuffering) => set({ isBuffering }),
  setPlaybackRate: (playbackRate) => set({ playbackRate }),
  setShowControls: (showControls) => set({ showControls }),
  reset: () => set(initialState),
}));
