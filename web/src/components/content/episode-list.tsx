'use client';

import Image from 'next/image';
import { useRouter } from 'next/navigation';

import { Play } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import { buildEpisodeContentId } from '@/lib/utils/episode-content-id';
import type { Episode } from '@/lib/types/catalog';

interface EpisodeListProps {
  episodes: Episode[];
  seriesId: string;
  seasonNumber: number;
  className?: string;
}

export default function EpisodeList({
  episodes,
  seriesId,
  seasonNumber,
  className,
}: EpisodeListProps) {
  const router = useRouter();
  const tb = useTranslations('browse');
  const tc = useTranslations('content');

  if (episodes.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        {tc('noEpisodes')}
      </p>
    );
  }

  const handlePlay = (episode: Episode) => {
    const contentId = buildEpisodeContentId(
      seriesId,
      seasonNumber,
      episode.episodeNumber,
    );
    const idx = episodes.findIndex((e) => e.episodeNumber === episode.episodeNumber);
    const next = idx >= 0 && idx < episodes.length - 1 ? episodes[idx + 1] : undefined;

    const params = new URLSearchParams({
      type: 'series',
      title: episode.title,
    });

    if (next) {
      params.set(
        'nextEpisodeId',
        buildEpisodeContentId(seriesId, seasonNumber, next.episodeNumber),
      );
      params.set('nextEpisodeTitle', next.title);
      params.set('nextSeason', String(seasonNumber));
      params.set('nextEpisodeNum', String(next.episodeNumber));
      if (next.thumbnailUrl) params.set('nextThumbnail', next.thumbnailUrl);
    }

    router.push(`/watch/${contentId}?${params.toString()}`);
  };

  return (
    <div className={cn('space-y-2', className)}>
      {episodes.map((episode) => (
        <button
          key={episode.episodeNumber}
          type="button"
          onClick={() => handlePlay(episode)}
          className={cn(
            'group flex w-full items-start gap-4 rounded-md p-3 text-left',
            'bg-muted/30 hover:bg-muted transition-colors',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
          )}
          aria-label={`${tb('episode', { number: episode.episodeNumber })} - ${episode.title}`}
        >
          {/* Thumbnail */}
          <div className="relative h-20 w-36 shrink-0 overflow-hidden rounded bg-muted">
            {episode.thumbnailUrl ? (
              <Image
                src={episode.thumbnailUrl}
                alt={episode.title}
                fill
                sizes="144px"
                className="object-cover"
              />
            ) : (
              <div className="absolute inset-0 bg-muted" />
            )}
            {/* Play overlay */}
            <div
              className={cn(
                'absolute inset-0 flex items-center justify-center',
                'bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity'
              )}
            >
              <Play className="h-8 w-8 fill-white text-white" />
            </div>
          </div>

          {/* Episode info */}
          <div className="min-w-0 flex-1">
            <div className="flex items-baseline justify-between gap-2">
              <p className="text-sm font-medium text-foreground">
                <span className="text-muted-foreground mr-2">
                  {tb('episode', { number: episode.episodeNumber })}
                </span>
                {episode.title}
              </p>
              <span className="shrink-0 text-xs text-muted-foreground">
                {episode.durationFormatted}
              </span>
            </div>
            {episode.description && (
              <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                {episode.description}
              </p>
            )}
          </div>
        </button>
      ))}
    </div>
  );
}
