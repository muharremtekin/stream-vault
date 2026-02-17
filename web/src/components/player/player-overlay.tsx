'use client';

import { useEffect, useRef, useState } from 'react';
import type { RefObject } from 'react';

import { ChevronLeft, RotateCcw } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { usePlayerStore } from '@/lib/stores/player-store';
import { cn } from '@/lib/utils/cn';

import type { NextEpisodeInfo } from '@/lib/types/streaming';

const COUNTDOWN_SECONDS = 10;

interface PlayerOverlayProps {
  isReady: boolean;
  isEnded: boolean;
  title?: string;
  onBack?: () => void;
  nextEpisode?: NextEpisodeInfo;
  videoRef: RefObject<HTMLVideoElement | null>;
  onReplay: () => void;
}

export function PlayerOverlay({
  isReady,
  isEnded,
  title,
  onBack,
  nextEpisode,
  videoRef: _videoRef,
  onReplay,
}: PlayerOverlayProps) {
  const t = useTranslations('player');
  const router = useRouter();
  const showControls = usePlayerStore((s) => s.showControls);
  const isBuffering = usePlayerStore((s) => s.isBuffering);
  const contentType = usePlayerStore((s) => s.contentType);

  const handleBack = () => {
    if (onBack) onBack();
    else router.back();
  };

  return (
    <>
      <TopBar
        isVisible={showControls && !isEnded}
        title={title}
        onBack={handleBack}
      />
      <BufferingSpinner isVisible={(!isReady || isBuffering) && !isEnded} label={t('buffering')} />
      {isEnded && (
        <EndScreen
          contentType={contentType}
          nextEpisode={nextEpisode}
          onReplay={onReplay}
          onBack={handleBack}
        />
      )}
    </>
  );
}

/* ------------------------------------------------------------------ */
/*  Top bar — title + back button                                     */
/* ------------------------------------------------------------------ */

interface TopBarProps {
  isVisible: boolean;
  title?: string;
  onBack: () => void;
}

function TopBar({ isVisible, title, onBack }: TopBarProps) {
  const t = useTranslations('player');
  return (
    <div
      className={cn(
        'pointer-events-none absolute inset-x-0 top-0 z-10 bg-gradient-to-b from-black/70 to-transparent px-4 pb-10 pt-4 transition-opacity duration-300',
        isVisible ? 'opacity-100' : 'opacity-0',
      )}
    >
      <div className="pointer-events-auto flex items-center gap-3 text-white">
        <button
          onClick={onBack}
          aria-label={t('back')}
          className="rounded-full p-1 transition-colors hover:bg-white/20 focus-visible:ring-2 focus-visible:ring-white"
        >
          <ChevronLeft className="h-6 w-6" />
        </button>
        {title && (
          <h1 className="truncate text-lg font-semibold">{title}</h1>
        )}
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Buffering spinner                                                  */
/* ------------------------------------------------------------------ */

interface BufferingSpinnerProps {
  isVisible: boolean;
  label: string;
}

function BufferingSpinner({ isVisible, label }: BufferingSpinnerProps) {
  if (!isVisible) return null;
  return (
    <div
      role="status"
      className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center"
    >
      <span className="sr-only">{label}</span>
      <div className="h-12 w-12 animate-spin rounded-full border-4 border-white border-t-transparent" />
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  End screen — next episode countdown or replay                     */
/* ------------------------------------------------------------------ */

interface EndScreenProps {
  contentType: string | null;
  nextEpisode?: NextEpisodeInfo;
  onReplay: () => void;
  onBack: () => void;
}

function EndScreen({ contentType, nextEpisode, onReplay, onBack }: EndScreenProps) {
  const t = useTranslations('player');
  const hasNextEpisode = contentType === 'series' && nextEpisode;

  if (hasNextEpisode) {
    return (
      <NextEpisodeCountdown
        nextEpisode={nextEpisode}
        onReplay={onReplay}
      />
    );
  }

  return (
    <div className="absolute inset-0 z-20 flex flex-col items-center justify-center gap-4 bg-black/80">
      <button
        onClick={onReplay}
        className="flex items-center gap-2 rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:ring-2 focus-visible:ring-white"
      >
        <RotateCcw className="h-4 w-4" />
        {t('playAgain')}
      </button>
      <button
        onClick={onBack}
        className="rounded-lg px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-white/20 focus-visible:ring-2 focus-visible:ring-white"
      >
        {t('goBack')}
      </button>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Next episode auto-play countdown                                   */
/* ------------------------------------------------------------------ */

interface NextEpisodeCountdownProps {
  nextEpisode: NextEpisodeInfo;
  onReplay: () => void;
}

function NextEpisodeCountdown({ nextEpisode, onReplay }: NextEpisodeCountdownProps) {
  const t = useTranslations('player');
  const router = useRouter();
  const [countdown, setCountdown] = useState(COUNTDOWN_SECONDS);
  const [isCancelled, setIsCancelled] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (isCancelled) return;
    intervalRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          if (intervalRef.current) clearInterval(intervalRef.current);
          router.push(`/watch/${nextEpisode.id}`);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [isCancelled, nextEpisode.id, router]);

  const handleCancel = () => {
    setIsCancelled(true);
    if (intervalRef.current) clearInterval(intervalRef.current);
  };

  const pct = isCancelled ? 0 : ((COUNTDOWN_SECONDS - countdown) / COUNTDOWN_SECONDS) * 100;

  return (
    <div className="absolute inset-0 z-20 flex flex-col items-center justify-center gap-5 bg-black/80">
      <p className="text-sm font-medium uppercase tracking-wider text-white/70">{t('upNext')}</p>
      <p className="max-w-md truncate text-lg font-semibold text-white">
        S{nextEpisode.seasonNumber} E{nextEpisode.episodeNumber} — {nextEpisode.title}
      </p>
      {!isCancelled && (
        <div className="h-1 w-48 overflow-hidden rounded-full bg-white/30">
          <div
            className="h-full bg-primary transition-[width] duration-1000 ease-linear"
            style={{ width: `${pct}%` }} /* dynamic countdown progress — inline required */
          />
        </div>
      )}
      <p className="text-sm text-white/60">
        {isCancelled ? t('nextEpisode') : t('skipCountdown', { seconds: countdown })}
      </p>
      <div className="flex gap-3">
        <button
          onClick={() => router.push(`/watch/${nextEpisode.id}`)}
          className="rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:ring-2 focus-visible:ring-white"
        >
          {t('nextEpisode')}
        </button>
        {!isCancelled ? (
          <button
            onClick={handleCancel}
            className="rounded-lg px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-white/20 focus-visible:ring-2 focus-visible:ring-white"
          >
            {t('cancelAutoPlay')}
          </button>
        ) : (
          <button
            onClick={onReplay}
            className="flex items-center gap-2 rounded-lg px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-white/20 focus-visible:ring-2 focus-visible:ring-white"
          >
            <RotateCcw className="h-4 w-4" />
            {t('playAgain')}
          </button>
        )}
      </div>
    </div>
  );
}
