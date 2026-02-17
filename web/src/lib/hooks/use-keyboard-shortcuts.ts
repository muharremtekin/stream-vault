'use client';

import { useEffect } from 'react';
import type { RefObject } from 'react';

import { usePlayerStore } from '@/lib/stores/player-store';

const SEEK_STEP = 10;
const VOLUME_STEP = 0.1;

const INTERACTIVE_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);

export function useKeyboardShortcuts(
  videoRef: RefObject<HTMLVideoElement | null>,
  containerRef: RefObject<HTMLDivElement | null>,
) {
  const setVolume = usePlayerStore((s) => s.setVolume);
  const setIsMuted = usePlayerStore((s) => s.setIsMuted);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (INTERACTIVE_TAGS.has(target.tagName) || target.isContentEditable) return;

      const video = videoRef.current;
      if (!video) return;

      switch (e.key) {
        case ' ': {
          e.preventDefault();
          if (video.paused) void video.play();
          else video.pause();
          break;
        }
        case 'f':
        case 'F': {
          e.preventDefault();
          const container = containerRef.current;
          if (!container) break;
          if (document.fullscreenElement) void document.exitFullscreen();
          else void container.requestFullscreen();
          break;
        }
        case 'ArrowLeft': {
          e.preventDefault();
          video.currentTime = Math.max(0, video.currentTime - SEEK_STEP);
          break;
        }
        case 'ArrowRight': {
          e.preventDefault();
          video.currentTime = Math.min(video.duration || 0, video.currentTime + SEEK_STEP);
          break;
        }
        case 'm':
        case 'M': {
          e.preventDefault();
          const muted = !video.muted;
          video.muted = muted;
          setIsMuted(muted);
          break;
        }
        case 'ArrowUp': {
          e.preventDefault();
          const volUp = Math.min(1, video.volume + VOLUME_STEP);
          video.volume = volUp;
          video.muted = false;
          setVolume(volUp);
          setIsMuted(false);
          break;
        }
        case 'ArrowDown': {
          e.preventDefault();
          const volDown = Math.max(0, video.volume - VOLUME_STEP);
          video.volume = volDown;
          if (volDown === 0) {
            video.muted = true;
            setIsMuted(true);
          }
          setVolume(volDown);
          break;
        }
        default:
          return;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [videoRef, containerRef, setVolume, setIsMuted]);
}
