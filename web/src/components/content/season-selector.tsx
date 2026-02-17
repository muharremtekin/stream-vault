'use client';

import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import type { Season } from '@/lib/types/catalog';

interface SeasonSelectorProps {
  seasons: Season[];
  selectedSeason: number;
  onSeasonChange: (seasonNumber: number) => void;
  className?: string;
}

export default function SeasonSelector({
  seasons,
  selectedSeason,
  onSeasonChange,
  className,
}: SeasonSelectorProps) {
  const tb = useTranslations('browse');
  const tc = useTranslations('content');

  if (seasons.length === 0) return null;

  return (
    <div className={cn('space-y-3', className)}>
      <h2 className="text-xl font-semibold text-foreground">{tc('selectSeason')}</h2>
      <div
        className={cn(
          'flex gap-2 overflow-x-auto pb-2',
          '[scrollbar-width:none] [&::-webkit-scrollbar]:hidden'
        )}
        role="tablist"
        aria-label={tc('selectSeason')}
      >
        {seasons.map((season) => {
          const isActive = season.seasonNumber === selectedSeason;
          return (
            <button
              key={season.seasonNumber}
              type="button"
              role="tab"
              aria-selected={isActive}
              onClick={() => onSeasonChange(season.seasonNumber)}
              className={cn(
                'shrink-0 rounded-md px-4 py-2 text-sm font-medium transition-colors',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80 hover:text-foreground'
              )}
            >
              <span className="block">{tb('season', { number: season.seasonNumber })}</span>
              <span className="block text-xs opacity-75">
                {tc('episodeCount', { count: season.episodeCount })}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
