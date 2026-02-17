'use client';

import { useRef, useState, useCallback } from 'react';

import Image from 'next/image';
import { useRouter } from 'next/navigation';

import ContentCardHover from '@/components/browse/content-card-hover';
import TrendingBadge from '@/components/browse/trending-badge';
import { cn } from '@/lib/utils/cn';
import type { ContentCardItem } from '@/lib/types/browse';

interface ContentCardProps {
  item: ContentCardItem;
  showRank?: boolean;
}

const HOVER_DELAY_MS = 300;

export default function ContentCard({ item, showRank = false }: ContentCardProps) {
  const router = useRouter();
  const cardRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [isHovered, setIsHovered] = useState(false);

  const handleMouseEnter = useCallback(() => {
    timerRef.current = setTimeout(() => setIsHovered(true), HOVER_DELAY_MS);
  }, []);

  const handleMouseLeave = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setIsHovered(false);
  }, []);

  const handleClick = () => {
    const path = item.contentType === 'Series'
      ? `/series/${item.contentId}`
      : `/movie/${item.contentId}`;
    router.push(path);
  };

  const hasProgress = item.percentage != null && item.percentage > 0;

  return (
    <div
      ref={cardRef}
      className="relative shrink-0 w-32 sm:w-36 md:w-44 overflow-visible"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {/* Base card */}
      <button
        type="button"
        onClick={handleClick}
        className={cn(
          'relative w-full rounded-md overflow-hidden cursor-pointer',
          'transition-transform duration-200',
          isHovered ? 'scale-105' : 'scale-100',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
        )}
      >
        {/* Poster image */}
        <div className="aspect-[2/3] w-full bg-muted">
          {item.thumbnailUrl ? (
            <Image
              src={item.thumbnailUrl}
              alt={item.title}
              fill
              sizes="(max-width: 640px) 128px, (max-width: 768px) 144px, 176px"
              className="object-cover"
            />
          ) : (
            <div className="absolute inset-0 flex items-end bg-gradient-to-b from-primary/10 to-muted p-2">
              <p className="text-xs text-muted-foreground line-clamp-2 font-medium">
                {item.title}
              </p>
            </div>
          )}
        </div>

        {/* Continue-watching progress bar */}
        {hasProgress && (
          <div className="absolute bottom-0 left-0 right-0 h-1 bg-white/20">
            {/* dynamic width — inline style required for percentage value */}
            <div
              className="h-full bg-primary"
              style={{ width: `${item.percentage}%` }} /* dynamic progress value */
            />
          </div>
        )}

        {/* Trending rank badge */}
        {showRank && item.rank != null && (
          <TrendingBadge rank={item.rank} />
        )}
      </button>

      {/* Hover overlay */}
      {isHovered && (
        <ContentCardHover
          item={item}
          cardRef={cardRef}
          onClose={handleMouseLeave}
        />
      )}
    </div>
  );
}
