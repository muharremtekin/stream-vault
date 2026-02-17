'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { RefObject } from 'react';

import type Hls from 'hls.js';
import { Maximize, Minimize, Pause, Play, Volume2, VolumeX } from 'lucide-react';
import dynamic from 'next/dynamic';
import { useTranslations } from 'next-intl';

import { usePlayerStore } from '@/lib/stores/player-store';
import { cn } from '@/lib/utils/cn';

const QualitySelector = dynamic(
  () => import('@/components/player/quality-selector').then((m) => m.QualitySelector),
  { ssr: false },
);

interface PlayerControlsProps {
  videoRef: RefObject<HTMLVideoElement | null>;
  containerRef: RefObject<HTMLDivElement | null>;
  hlsRef?: RefObject<Hls | null>;
}

interface SeekBarProps {
  currentTime: number;
  duration: number;
  onSeek: (e: React.ChangeEvent<HTMLInputElement>) => void;
  ariaLabel: string;
}

function formatPlayerTime(seconds: number): string {
  if (!isFinite(seconds) || seconds < 0) return '0:00';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) {
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  }
  return `${m}:${String(s).padStart(2, '0')}`;
}

function SeekBar({ currentTime, duration, onSeek, ariaLabel }: SeekBarProps) {
  const pct = duration > 0 ? (currentTime / duration) * 100 : 0;
  return (
    <div className="relative mb-3 h-1">
      <div className="absolute inset-0 rounded-full bg-white/30" />
      <div
        className="absolute inset-y-0 left-0 rounded-full bg-primary"
        style={{ width: `${pct}%` }} /* dynamic progress — inline required */
      />
      <input
        type="range"
        min={0}
        max={duration || 1}
        value={currentTime}
        step={0.5}
        onChange={onSeek}
        className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
        aria-label={ariaLabel}
      />
    </div>
  );
}

export function PlayerControls({ videoRef, containerRef, hlsRef }: PlayerControlsProps) {
  const t = useTranslations('player');
  const [isVisible, setIsVisible] = useState(true);
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const currentTime = usePlayerStore((s) => s.currentTime);
  const duration = usePlayerStore((s) => s.duration);
  const volume = usePlayerStore((s) => s.volume);
  const isMuted = usePlayerStore((s) => s.isMuted);
  const isFullscreen = usePlayerStore((s) => s.isFullscreen);
  const setVolume = usePlayerStore((s) => s.setVolume);
  const setIsMuted = usePlayerStore((s) => s.setIsMuted);
  const setIsFullscreen = usePlayerStore((s) => s.setIsFullscreen);

  const resetHideTimer = useCallback(() => {
    setIsVisible(true);
    if (hideTimerRef.current) clearTimeout(hideTimerRef.current);
    hideTimerRef.current = setTimeout(() => setIsVisible(false), 3000);
  }, []);

  useEffect(() => {
    resetHideTimer();
    return () => {
      if (hideTimerRef.current) clearTimeout(hideTimerRef.current);
    };
  }, [resetHideTimer]);

  useEffect(() => {
    const handleFsChange = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener('fullscreenchange', handleFsChange);
    return () => document.removeEventListener('fullscreenchange', handleFsChange);
  }, [setIsFullscreen]);

  const handlePlayPause = () => {
    const video = videoRef.current;
    if (!video) return;
    if (video.paused) void video.play();
    else video.pause();
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const video = videoRef.current;
    if (!video) return;
    video.currentTime = Number(e.target.value);
  };

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const video = videoRef.current;
    if (!video) return;
    const vol = Number(e.target.value);
    video.volume = vol;
    video.muted = vol === 0;
    setVolume(vol);
    setIsMuted(vol === 0);
  };

  const handleMuteToggle = () => {
    const video = videoRef.current;
    if (!video) return;
    const muted = !video.muted;
    video.muted = muted;
    setIsMuted(muted);
  };

  const handleFullscreen = async () => {
    const container = containerRef.current;
    if (!container) return;
    if (document.fullscreenElement) {
      await document.exitFullscreen();
    } else {
      await container.requestFullscreen();
    }
  };

  return (
    <div className="absolute inset-0" onMouseMove={resetHideTimer}>
      <div
        className={cn(
          'absolute inset-0 bg-gradient-to-t from-black/80 to-transparent transition-opacity duration-300',
          isVisible ? 'opacity-100' : 'opacity-0',
        )}
      />
      <div
        className={cn(
          'absolute bottom-0 left-0 right-0 px-4 pb-3 transition-opacity duration-300',
          isVisible ? 'opacity-100' : 'opacity-0 pointer-events-none',
        )}
      >
        <SeekBar
          currentTime={currentTime}
          duration={duration}
          onSeek={handleSeek}
          ariaLabel={t('seekBar')}
        />
        <div className="flex items-center gap-3 text-white">
          <button
            onClick={handlePlayPause}
            aria-label={isPlaying ? t('pause') : t('play')}
            className="rounded focus-visible:ring-2 focus-visible:ring-white"
          >
            {isPlaying ? <Pause className="h-5 w-5" /> : <Play className="h-5 w-5" />}
          </button>
          <span className="text-sm tabular-nums">
            {formatPlayerTime(currentTime)} / {formatPlayerTime(duration)}
          </span>
          <div className="flex-1" />
          <button
            onClick={handleMuteToggle}
            aria-label={isMuted ? t('unmute') : t('mute')}
            className="rounded focus-visible:ring-2 focus-visible:ring-white"
          >
            {isMuted || volume === 0 ? (
              <VolumeX className="h-5 w-5" />
            ) : (
              <Volume2 className="h-5 w-5" />
            )}
          </button>
          <input
            type="range"
            min={0}
            max={1}
            step={0.05}
            value={isMuted ? 0 : volume}
            onChange={handleVolumeChange}
            className="w-20 cursor-pointer accent-white"
            aria-label={t('volume')}
          />
          {hlsRef && <QualitySelector hlsRef={hlsRef} />}
          <button
            onClick={() => void handleFullscreen()}
            aria-label={isFullscreen ? t('exitFullscreen') : t('fullscreen')}
            className="rounded focus-visible:ring-2 focus-visible:ring-white"
          >
            {isFullscreen ? (
              <Minimize className="h-5 w-5" />
            ) : (
              <Maximize className="h-5 w-5" />
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
