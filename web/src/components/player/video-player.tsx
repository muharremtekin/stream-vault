'use client';

import { useCallback, useEffect, useRef } from 'react';
import type { Ref } from 'react';

import { useQuery } from '@tanstack/react-query';
import { useTranslations } from 'next-intl';

import { getStreamingInfo } from '@/lib/api/streaming';
import { useHlsPlayer } from '@/lib/hooks/use-hls-player';
import { usePlayerStore } from '@/lib/stores/player-store';
import { cn } from '@/lib/utils/cn';
import { PlayerControls } from '@/components/player/player-controls';

import type { StreamingInfo } from '@/lib/types/streaming';

interface VideoPlayerProps {
  contentId: string;
  contentType?: string;
  className?: string;
  ref?: Ref<HTMLVideoElement | null>;
}

export function VideoPlayer({
  contentId,
  contentType = 'movie',
  className,
  ref,
}: VideoPlayerProps) {
  const t = useTranslations('player');
  const internalRef = useRef<HTMLVideoElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const setContent = usePlayerStore((s) => s.setContent);
  const reset = usePlayerStore((s) => s.reset);
  const setIsPlaying = usePlayerStore((s) => s.setIsPlaying);
  const setCurrentTime = usePlayerStore((s) => s.setCurrentTime);
  const setDuration = usePlayerStore((s) => s.setDuration);
  const setIsBuffering = usePlayerStore((s) => s.setIsBuffering);
  const setAvailableQualities = usePlayerStore((s) => s.setAvailableQualities);

  // Merge internal ref with optional external ref (React 19 — no forwardRef needed)
  const mergedRef = useCallback(
    (node: HTMLVideoElement | null) => {
      internalRef.current = node;
      if (typeof ref === 'function') {
        ref(node);
      } else if (ref !== null && ref !== undefined) {
        (ref as { current: HTMLVideoElement | null }).current = node;
      }
    },
    [ref],
  );

  const { data: streamingInfo, isLoading, error: queryError } =
    useQuery<StreamingInfo>({
      queryKey: ['streaming', 'info', contentId],
      queryFn: () => getStreamingInfo(contentId),
      staleTime: 5 * 60_000,
      enabled: !!contentId,
    });

  const manifestUrl =
    streamingInfo?.videoStatus === 'Ready' ? streamingInfo.manifestUrl : null;

  const { isReady, error: hlsError } = useHlsPlayer(internalRef, {
    manifestUrl,
    isEnabled: !!manifestUrl,
  });

  // Initialise player store content on mount, reset on unmount
  useEffect(() => {
    setContent(contentId, contentType);
    return () => {
      reset();
    };
  }, [contentId, contentType, setContent, reset]);

  // Sync available qualities to player store
  useEffect(() => {
    const qualities =
      streamingInfo?.availableQualities?.map((q) => q.label) ?? [];
    setAvailableQualities(qualities);
  }, [streamingInfo, setAvailableQualities]);

  if (isLoading) {
    return (
      <div
        role="status"
        className={cn(
          'flex items-center justify-center bg-black',
          className,
        )}
      >
        <span className="sr-only">{t('loading')}</span>
        <div className="h-10 w-10 animate-spin rounded-full border-4 border-white border-t-transparent" />
      </div>
    );
  }

  if (queryError || hlsError === 'fatal' || hlsError === 'unsupported') {
    return (
      <div
        role="alert"
        className={cn(
          'flex items-center justify-center bg-black text-white',
          className,
        )}
      >
        <p>{hlsError === 'unsupported' ? t('unsupported') : t('error')}</p>
      </div>
    );
  }

  if (streamingInfo?.videoStatus === 'Processing') {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-black text-center text-white',
          className,
        )}
      >
        <p>{t('encoding')}</p>
      </div>
    );
  }

  if (streamingInfo && streamingInfo.videoStatus !== 'Ready') {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-black text-white',
          className,
        )}
      >
        <p>{t('notReady')}</p>
      </div>
    );
  }

  return (
    <div ref={containerRef} className={cn('relative bg-black', className)}>
      <video
        ref={mergedRef}
        className="h-full w-full"
        playsInline
        aria-label={t('videoLabel')}
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onEnded={() => setIsPlaying(false)}
        onWaiting={() => setIsBuffering(true)}
        onCanPlay={() => setIsBuffering(false)}
        onTimeUpdate={(e) => setCurrentTime(e.currentTarget.currentTime)}
        onDurationChange={(e) => setDuration(e.currentTarget.duration)}
      />
      {!isReady && (
        <div
          role="status"
          className="absolute inset-0 flex items-center justify-center bg-black"
        >
          <span className="sr-only">{t('loading')}</span>
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-white border-t-transparent" />
        </div>
      )}
      {isReady && (
        <PlayerControls videoRef={internalRef} containerRef={containerRef} />
      )}
    </div>
  );
}
