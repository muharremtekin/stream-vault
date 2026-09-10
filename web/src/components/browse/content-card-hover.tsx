'use client';

import { useEffect, useState } from 'react';

import { Info, Play, Star, ThumbsUp } from 'lucide-react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { Button } from '@/components/ui/button';
import GenreTags from '@/components/browse/genre-tags';
import { useWatchlist } from '@/lib/hooks/use-watchlist';
import type { ContentCardItem } from '@/lib/types/browse';
import { cn } from '@/lib/utils/cn';

interface ContentCardHoverProps {
  item: ContentCardItem;
  cardRef: React.RefObject<HTMLDivElement | null>;
  onClose: () => void;
}

type EdgeAlign = 'left' | 'center' | 'right';

function resolveAlignment(cardRef: React.RefObject<HTMLDivElement | null>): EdgeAlign {
  if (!cardRef.current) return 'center';
  const { left, right } = cardRef.current.getBoundingClientRect();
  if (left < 200) return 'left';
  if (right > window.innerWidth - 200) return 'right';
  return 'center';
}

export default function ContentCardHover({ item, cardRef, onClose }: ContentCardHoverProps) {
  const t = useTranslations('browse');
  const router = useRouter();
  const { isInWatchlist, add, remove, getWatchlistItemId } = useWatchlist();

  const [align, setAlign] = useState<EdgeAlign>('center');
  const inList = isInWatchlist(item.contentId);
  const contentPath = item.contentType === 'Series'
    ? `/series/${item.contentId}`
    : `/movie/${item.contentId}`;
  const playbackPath = item.playbackHref ?? `/watch/${item.contentId}`;

  useEffect(() => {
    setAlign(resolveAlignment(cardRef));
  }, [cardRef]);

  const handleWatchlist = () => {
    if (inList) {
      const wid = getWatchlistItemId(item.contentId);
      if (wid) remove(wid);
    } else {
      add({ contentId: item.contentId, contentType: item.contentType === 'Series' ? 'Series' : 'Movie' });
    }
  };

  const positionClass = {
    left: 'left-0',
    right: 'right-0',
    center: 'left-1/2 -translate-x-1/2',
  }[align];

  return (
    <div
      onMouseLeave={onClose}
      className={cn(
        'absolute bottom-0 z-50 w-72 rounded-md overflow-hidden',
        'bg-card shadow-2xl border border-white/10',
        'animate-in fade-in zoom-in-95 duration-200',
        positionClass
      )}
    >
      {/* Thumbnail */}
      <div className="relative aspect-video w-full bg-muted">
        {item.thumbnailUrl ? (
          <Image src={item.thumbnailUrl} alt={item.title} fill className="object-cover" />
        ) : (
          <div className="absolute inset-0 bg-gradient-to-br from-primary/20 to-muted" />
        )}
      </div>

      {/* Info section */}
      <div className="p-3 space-y-2">
        {/* Action buttons */}
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="primary"
            onClick={() => router.push(playbackPath)}
            aria-label={t('play')}
          >
            <Play className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="secondary" onClick={handleWatchlist}>
            {inList ? t('removeFromList') : t('addToList')}
          </Button>
          <Button size="sm" variant="ghost" aria-label={t('like')}>
            <ThumbsUp className="h-4 w-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => router.push(contentPath)}
            aria-label={t('moreInfo')}
            className="ml-auto"
          >
            <Info className="h-4 w-4" />
          </Button>
        </div>

        {/* Title & meta */}
        <p className="font-semibold text-sm text-foreground line-clamp-1">{item.title}</p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          {item.averageRating != null && (
            <span className="flex items-center gap-0.5 text-green-400">
              <Star className="h-3 w-3 fill-current" />
              {item.averageRating.toFixed(1)}
            </span>
          )}
          {item.releaseYear && <span>{item.releaseYear}</span>}
          {item.durationMinutes && <span>{item.durationMinutes}m</span>}
          {item.totalSeasons && <span>{item.totalSeasons}s</span>}
        </div>

        {/* Genre tags */}
        {item.genres && item.genres.length > 0 && (
          <GenreTags genres={item.genres.slice(0, 3)} />
        )}
      </div>
    </div>
  );
}
