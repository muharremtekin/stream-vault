import type { Metadata } from 'next';

import type { NextEpisodeInfo } from '@/lib/types/streaming';
import { WatchPlayerClient } from './watch-player-client';

interface WatchPageProps {
  params: Promise<{ contentId: string }>;
  searchParams: Promise<{
    type?: string;
    title?: string;
    nextEpisodeId?: string;
    nextEpisodeTitle?: string;
    nextSeason?: string;
    nextEpisodeNum?: string;
    nextThumbnail?: string;
  }>;
}

export async function generateMetadata(props: WatchPageProps): Promise<Metadata> {
  const { title } = await props.searchParams;
  return {
    title: title ? `${title} — StreamVault` : 'Watch — StreamVault',
  };
}

export default async function WatchPage(props: WatchPageProps) {
  const { contentId } = await props.params;
  const {
    type,
    title,
    nextEpisodeId,
    nextEpisodeTitle,
    nextSeason,
    nextEpisodeNum,
    nextThumbnail,
  } = await props.searchParams;

  const contentType: 'movie' | 'series' = type === 'series' ? 'series' : 'movie';

  const nextEpisode: NextEpisodeInfo | undefined = nextEpisodeId
    ? {
        id: nextEpisodeId,
        title: nextEpisodeTitle ?? '',
        seasonNumber: parseInt(nextSeason ?? '1', 10),
        episodeNumber: parseInt(nextEpisodeNum ?? '1', 10),
        thumbnailUrl: nextThumbnail ?? null,
      }
    : undefined;

  return (
    <WatchPlayerClient
      contentId={contentId}
      contentType={contentType}
      title={title}
      nextEpisode={nextEpisode}
    />
  );
}
