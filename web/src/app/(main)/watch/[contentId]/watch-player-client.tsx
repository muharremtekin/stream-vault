'use client';

import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';

import type { NextEpisodeInfo } from '@/lib/types/streaming';

const VideoPlayer = dynamic(
  () => import('@/components/player/video-player').then((m) => ({ default: m.VideoPlayer })),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-full w-full items-center justify-center bg-black">
        <div className="h-10 w-10 animate-spin rounded-full border-4 border-white border-t-transparent" />
      </div>
    ),
  },
);

interface WatchPlayerClientProps {
  contentId: string;
  contentType: 'movie' | 'series';
  title?: string;
  nextEpisode?: NextEpisodeInfo;
}

export function WatchPlayerClient({
  contentId,
  contentType,
  title,
  nextEpisode,
}: WatchPlayerClientProps) {
  const router = useRouter();

  return (
    <div className="fixed inset-0 z-50 bg-black">
      <VideoPlayer
        contentId={contentId}
        contentType={contentType}
        title={title}
        onBack={() => router.back()}
        nextEpisode={nextEpisode}
        className="h-full w-full"
      />
    </div>
  );
}
