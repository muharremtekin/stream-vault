'use client';

import { useEffect, useRef, useState } from 'react';
import type { RefObject } from 'react';

import Hls from 'hls.js';
import { Settings } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useAuthStore } from '@/lib/stores/auth-store';
import { usePlayerStore } from '@/lib/stores/player-store';
import { cn } from '@/lib/utils/cn';

interface HlsLevel {
  index: number;
  height: number;
  label: string;
}

interface QualitySelectorProps {
  hlsRef: RefObject<Hls | null>;
}

const TIER_MAX_HEIGHT: Record<string, number> = {
  Free: 480,
  Basic: 720,
  Standard: 1080,
  Premium: 2160,
  Admin: 2160,
};

function heightToLabel(height: number): string {
  if (height >= 2160) return '4K';
  if (height >= 1440) return '1440p';
  if (height >= 1080) return '1080p';
  if (height >= 720) return '720p';
  if (height >= 480) return '480p';
  if (height >= 360) return '360p';
  return `${height}p`;
}

export function QualitySelector({ hlsRef }: QualitySelectorProps) {
  const t = useTranslations('player');
  const [isOpen, setIsOpen] = useState(false);
  const [hlsLevels, setHlsLevels] = useState<HlsLevel[]>([]);
  const [selectedLevel, setSelectedLevel] = useState<number>(-1);
  const menuRef = useRef<HTMLDivElement | null>(null);

  const subscriptionTier = useAuthStore((s) => s.subscriptionTier);
  const setCurrentQuality = usePlayerStore((s) => s.setCurrentQuality);

  const maxHeight = TIER_MAX_HEIGHT[subscriptionTier ?? 'Free'] ?? 480;

  useEffect(() => {
    const hls = hlsRef.current;
    if (!hls) return;

    const extractLevels = () => {
      const levels: HlsLevel[] = hls.levels.map((level, index) => ({
        index,
        height: level.height,
        label: heightToLabel(level.height),
      }));
      setHlsLevels(levels);
      setSelectedLevel(hls.currentLevel);
    };

    const onLevelSwitched = () => {
      // Track currently playing level when in auto mode
      if (selectedLevel === -1) {
        const playing = hls.levels[hls.currentLevel];
        if (playing) setCurrentQuality(heightToLabel(playing.height));
      }
    };

    hls.on(Hls.Events.MANIFEST_PARSED, extractLevels);
    hls.on(Hls.Events.LEVEL_SWITCHED, onLevelSwitched);

    // Manifest may already be parsed by the time this component mounts
    if (hls.levels.length > 0) {
      extractLevels();
    }

    return () => {
      hls.off(Hls.Events.MANIFEST_PARSED, extractLevels);
      hls.off(Hls.Events.LEVEL_SWITCHED, onLevelSwitched);
    };
  }, [hlsRef, setCurrentQuality, selectedLevel]);

  // Close dropdown on outside click
  useEffect(() => {
    if (!isOpen) return;
    const handleOutsideClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  const handleSelect = (levelIndex: number) => {
    const hls = hlsRef.current;
    if (!hls) return;
    hls.currentLevel = levelIndex;
    setSelectedLevel(levelIndex);
    if (levelIndex === -1) {
      setCurrentQuality('Auto');
    } else {
      const level = hls.levels[levelIndex];
      if (level) setCurrentQuality(heightToLabel(level.height));
    }
    setIsOpen(false);
  };

  if (hlsLevels.length === 0) return null;

  const currentLabel =
    selectedLevel === -1
      ? t('qualityAuto')
      : (hlsLevels.find((l) => l.index === selectedLevel)?.label ?? t('qualityAuto'));

  return (
    <div ref={menuRef} className="relative">
      <button
        onClick={() => setIsOpen((prev) => !prev)}
        aria-label={`${t('qualitySelect')}: ${currentLabel}`}
        aria-expanded={isOpen}
        aria-haspopup="true"
        className="rounded p-0.5 text-white focus-visible:ring-2 focus-visible:ring-white"
      >
        <Settings className="h-5 w-5" />
      </button>

      {isOpen && (
        <div
          role="menu"
          className="absolute bottom-8 right-0 z-50 min-w-32 rounded-md bg-black/90 py-1 text-sm text-white shadow-lg"
        >
          <p className="px-3 py-1 text-xs font-semibold uppercase tracking-wider text-white/60">
            {t('quality')}
          </p>

          <button
            role="menuitem"
            onClick={() => handleSelect(-1)}
            className={cn(
              'flex w-full items-center px-3 py-1.5 text-left hover:bg-white/10',
              selectedLevel === -1 && 'font-semibold text-primary',
            )}
          >
            {t('qualityAuto')}
          </button>

          {[...hlsLevels].reverse().map((level) => {
            const isDisabled = level.height > maxHeight;
            const isActive = level.index === selectedLevel;
            return (
              <button
                key={level.index}
                role="menuitem"
                onClick={() => { if (!isDisabled) handleSelect(level.index); }}
                disabled={isDisabled}
                aria-disabled={isDisabled}
                className={cn(
                  'flex w-full items-center justify-between px-3 py-1.5 text-left',
                  isActive && !isDisabled && 'font-semibold text-primary',
                  isDisabled
                    ? 'cursor-not-allowed text-white/30'
                    : 'hover:bg-white/10',
                )}
              >
                <span>{level.label}</span>
                {isDisabled && (
                  <span className="ml-2 rounded bg-white/10 px-1.5 py-0.5 text-xs">
                    {t('qualityUpgrade')}
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
