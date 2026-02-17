'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { Ref } from 'react';

import { useQuery } from '@tanstack/react-query';
import { useTranslations } from 'next-intl';

import { getStreamingInfo } from '@/lib/api/streaming';
import { useHlsPlayer } from '@/lib/hooks/use-hls-player';
import { useKeyboardShortcuts } from '@/lib/hooks/use-keyboard-shortcuts';
import { usePlayerStore } from '@/lib/stores/player-store';
import { cn } from '@/lib/utils/cn';
import { API_URL } from '@/lib/utils/constants';
import { PlayerControls } from '@/components/player/player-controls';
import { PlayerOverlay } from '@/components/player/player-overlay';
import { ProgressTracker } from '@/components/player/progress-tracker';

import type { NextEpisodeInfo, StreamingInfo } from '@/lib/types/streaming';

interface VideoPlayerProps {
  contentId: string;
  contentType?: string;
  title?: string;
  onBack?: () => void;
  nextEpisode?: NextEpisodeInfo;
  className?: string;
  ref?: Ref<HTMLVideoElement | null>;
}

export function VideoPlayer({
  contentId,
  contentType = 'movie',
  title,
  onBack,
  nextEpisode,
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
  const setShowControls = usePlayerStore((s) => s.setShowControls);
  const [isEnded, setIsEnded] = useState(false);
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleMouseMove = useCallback(() => {
    setShowControls(true);
    if (hideTimerRef.current) clearTimeout(hideTimerRef.current);
    hideTimerRef.current = setTimeout(() => setShowControls(false), 3000);
  }, [setShowControls]);

  useKeyboardShortcuts(internalRef, containerRef);

  const mergedRef = useCallback(
    (node: HTMLVideoElement | null) => {
      internalRef.current = node;
      if (typeof ref === 'function') ref(node);
      else if (ref !== null && ref !== undefined) {
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

  // Backend returns a path (e.g. /stream/{id}/manifest.m3u8) — prefix with gateway URL
  // so hls.js fetches from the API gateway, not the Next.js dev server.
  const rawManifestUrl = streamingInfo?.videoStatus === 'Ready' ? streamingInfo.manifestUrl : null;
  const manifestUrl = rawManifestUrl
    ? rawManifestUrl.startsWith('http')
      ? rawManifestUrl
      : `${API_URL}${rawManifestUrl}`
    : null;

  const { isReady, error: hlsError, hlsRef } = useHlsPlayer(internalRef, {
    manifestUrl,
    isEnabled: !!manifestUrl,
  });

  useEffect(() => {
    setContent(contentId, contentType);
    return () => { reset(); };
  }, [contentId, contentType, setContent, reset]);

  useEffect(() => {
    setAvailableQualities(streamingInfo?.availableQualities?.map((q) => q.label) ?? []);
  }, [streamingInfo, setAvailableQualities]);

  const handleReplay = useCallback(() => {
    const video = internalRef.current;
    if (!video) return;
    video.currentTime = 0;
    setIsEnded(false);
    void video.play();
  }, []);

  if (isLoading) {
    return (
      <div role="status" className={cn('flex items-center justify-center bg-black', className)}>
        <span className="sr-only">{t('loading')}</span>
        <div className="h-10 w-10 animate-spin rounded-full border-4 border-white border-t-transparent" />
      </div>
    );
  }

  if (queryError || hlsError === 'fatal' || hlsError === 'unsupported') {
    return (
      <div role="alert" className={cn('flex items-center justify-center bg-black text-white', className)}>
        <p>{hlsError === 'unsupported' ? t('unsupported') : t('error')}</p>
      </div>
    );
  }

  if (streamingInfo?.videoStatus === 'Processing') {
    return (
      <div className={cn('flex items-center justify-center bg-black text-center text-white', className)}>
        <p>{t('encoding')}</p>
      </div>
    );
  }

  if (streamingInfo && streamingInfo.videoStatus !== 'Ready') {
    return (
      <div className={cn('flex items-center justify-center bg-black text-white', className)}>
        <p>{t('notReady')}</p>
      </div>
    );
  }

  return (
    <div ref={containerRef} className={cn('relative bg-black', className)} onMouseMove={handleMouseMove}>
      <video
        ref={mergedRef}
        className="h-full w-full"
        playsInline
        aria-label={t('videoLabel')}
        onPlay={() => { setIsPlaying(true); setIsEnded(false); }}
        onPause={() => setIsPlaying(false)}
        onEnded={() => { setIsPlaying(false); setIsEnded(true); }}
        onWaiting={() => setIsBuffering(true)}
        onCanPlay={() => setIsBuffering(false)}
        onTimeUpdate={(e) => setCurrentTime(e.currentTarget.currentTime)}
        onDurationChange={(e) => setDuration(e.currentTarget.duration)}
      />
      <PlayerOverlay
        isReady={isReady}
        isEnded={isEnded}
        title={title}
        onBack={onBack}
        nextEpisode={nextEpisode}
        videoRef={internalRef}
        onReplay={handleReplay}
      />
      {isReady && (
        <>
          <ProgressTracker videoRef={internalRef} contentId={contentId} />
          <PlayerControls videoRef={internalRef} containerRef={containerRef} hlsRef={hlsRef} />
        </>
      )}
    </div>
  );
}
