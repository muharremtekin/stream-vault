'use client';

import { useState } from 'react';

import { Skeleton } from '@/components/ui/skeleton';
import EpisodeList from '@/components/content/episode-list';
import SeasonSelector from '@/components/content/season-selector';
import { useSeries } from '@/lib/hooks/use-content';

interface SeriesEpisodesSectionProps {
  seriesId: string;
}

export default function SeriesEpisodesSection({ seriesId }: SeriesEpisodesSectionProps) {
  const { data: series, isLoading } = useSeries(seriesId);
  const [selectedSeason, setSelectedSeason] = useState(1);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex gap-2">
          {Array.from({ length: 3 }).map((_, i) => (
            // eslint-disable-next-line react/no-array-index-key
            <Skeleton key={`season-skeleton-${i}`} variant="custom" className="h-14 w-24 rounded-md" />
          ))}
        </div>
        {Array.from({ length: 4 }).map((_, i) => (
          // eslint-disable-next-line react/no-array-index-key
          <Skeleton key={`ep-skeleton-${i}`} variant="custom" className="h-24 w-full rounded-md" />
        ))}
      </div>
    );
  }

  if (!series || series.seasons.length === 0) return null;

  const activeSeason = series.seasons.find(
    (s) => s.seasonNumber === selectedSeason
  ) ?? series.seasons[0];

  return (
    <div className="space-y-6">
      <SeasonSelector
        seasons={series.seasons}
        selectedSeason={activeSeason.seasonNumber}
        onSeasonChange={setSelectedSeason}
      />
      <EpisodeList
        episodes={activeSeason.episodes}
        seriesId={seriesId}
        seasonNumber={activeSeason.seasonNumber}
      />
    </div>
  );
}
