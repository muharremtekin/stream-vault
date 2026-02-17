'use client';

import Image from 'next/image';
import { useRouter } from 'next/navigation';

import { Play } from 'lucide-react';
import { useTranslations } from 'next-intl';

import AddToListButton from '@/components/content/add-to-list-button';
import MaturityBadge from '@/components/content/maturity-badge';
import RatingStars from '@/components/content/rating-stars';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useContent } from '@/lib/hooks/use-content';
import { cn } from '@/lib/utils/cn';
import type { Movie } from '@/lib/types/catalog';
import type { ContentType } from '@/lib/types/common';

interface ContentHeroProps {
  id: string;
  contentType: 'movie' | 'series';
}

function isMovie(data: unknown): data is Movie {
  return typeof data === 'object' && data !== null && 'durationMinutes' in data;
}

export default function ContentHero({ id, contentType }: ContentHeroProps) {
  const router = useRouter();
  const t = useTranslations('browse');
  const tc = useTranslations('content');
  const { data, isLoading } = useContent(id, contentType);

  if (isLoading || !data) {
    return (
      <div className="relative min-h-[70vh] w-full bg-muted">
        <Skeleton variant="custom" className="absolute inset-0" />
        <div className="absolute bottom-0 left-0 right-0 p-8 sm:p-12 lg:p-16 space-y-4">
          <Skeleton variant="text" className="h-4 w-24" />
          <Skeleton variant="text" className="h-12 w-2/3" />
          <Skeleton variant="text" className="h-4 w-48" />
          <div className="flex gap-3 pt-2">
            <Skeleton variant="custom" className="h-11 w-28 rounded-md" />
            <Skeleton variant="custom" className="h-11 w-36 rounded-md" />
          </div>
        </div>
      </div>
    );
  }

  const apiContentType: ContentType = contentType === 'movie' ? 'Movie' : 'Series';
  const movie = isMovie(data) ? data : null;
  const metaInfo = movie
    ? [data.releaseYear, movie.durationFormatted].filter(Boolean).join(' · ')
    : [data.releaseYear, `${'totalSeasons' in data ? data.totalSeasons : ''} ${t('seasons')}`].filter(Boolean).join(' · ');

  const backgroundUrl = data.bannerUrl || data.thumbnailUrl;

  return (
    <div className="relative min-h-[70vh] w-full overflow-hidden">
      {/* Background image */}
      {backgroundUrl && (
        <Image
          src={backgroundUrl}
          alt={data.title}
          fill
          priority
          sizes="100vw"
          className="object-cover object-center"
        />
      )}

      {/* Gradients */}
      <div className="absolute inset-0 bg-gradient-to-r from-background/90 via-background/50 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent" />

      {/* Content */}
      <div className="absolute bottom-0 left-0 right-0 px-4 pb-12 sm:px-8 lg:px-12 sm:pb-16">
        <div className="max-w-xl space-y-3">
          <MaturityBadge rating={data.maturityRating} />

          <h1 className="text-4xl font-bold text-foreground sm:text-5xl leading-tight">
            {data.title}
          </h1>

          <div className={cn('flex items-center gap-3 text-sm text-muted-foreground')}>
            <RatingStars rating={data.averageRating} totalRatings={data.ratingCount} size="sm" />
            <span>{metaInfo}</span>
          </div>

          {data.genres.length > 0 && (
            <p className="text-sm text-muted-foreground">
              {data.genres.slice(0, 3).join(' · ')}
            </p>
          )}

          <div className="flex flex-wrap items-center gap-3 pt-2">
            <Button
              onClick={() => router.push(`/watch/${id}`)}
              className="gap-2"
            >
              <Play className="h-4 w-4 fill-current" />
              {t('play')}
            </Button>
            <AddToListButton
              contentId={id}
              contentType={apiContentType}
              variant="button"
            />
          </div>

          {!movie && 'status' in data && data.status === 'Ended' && (
            <p className="text-xs text-muted-foreground">{tc('endedYear')}</p>
          )}
        </div>
      </div>
    </div>
  );
}
