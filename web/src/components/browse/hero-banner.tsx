'use client';

import { Info, List, Play } from 'lucide-react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useFeaturedMovie } from '@/lib/hooks/use-browse';
import { useWatchlist } from '@/lib/hooks/use-watchlist';
import type { Movie } from '@/lib/types/catalog';
import { cn } from '@/lib/utils/cn';

interface HeroBannerProps {
  movie: Movie;
}

export function HeroBanner({ movie }: HeroBannerProps) {
  const t = useTranslations('browse');
  const router = useRouter();
  const { isInWatchlist, add, remove, getWatchlistItemId } = useWatchlist();

  const inList = isInWatchlist(movie.id);

  const handleWatchlist = () => {
    if (inList) {
      const wid = getWatchlistItemId(movie.id);
      if (wid) remove(wid);
    } else {
      add({ contentId: movie.id, contentType: 'Movie' });
    }
  };

  return (
    <div className="relative h-[80vh] min-h-[500px] w-full overflow-hidden">
      {/* Background image */}
      <Image
        src={movie.bannerUrl || movie.thumbnailUrl}
        alt={movie.title}
        fill
        priority
        sizes="100vw"
        className="object-cover"
      />

      {/* Gradient overlays */}
      <div className="absolute inset-0 bg-gradient-to-r from-background/90 via-background/40 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent" />

      {/* Content */}
      <div className="absolute bottom-24 left-4 sm:left-8 lg:left-12 max-w-lg space-y-4">
        {/* Meta badges */}
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span className="rounded border border-muted-foreground px-1.5 py-0.5 text-xs">
            {movie.maturityRating}
          </span>
          <span>{movie.releaseYear}</span>
          <span>{movie.durationFormatted}</span>
          {movie.averageRating > 0 && (
            <span className="text-green-400">★ {movie.averageRating.toFixed(1)}</span>
          )}
        </div>

        {/* Title */}
        <h1 className="text-3xl font-bold text-foreground sm:text-4xl lg:text-5xl drop-shadow-lg">
          {movie.title}
        </h1>

        {/* Description */}
        <p className="text-sm text-foreground/80 line-clamp-3 sm:text-base">
          {movie.description}
        </p>

        {/* Action buttons */}
        <div className="flex flex-wrap items-center gap-3">
          <Button
            variant="primary"
            size="lg"
            onClick={() => router.push(`/watch/${movie.id}`)}
          >
            <Play className="mr-2 h-5 w-5 fill-current" />
            {t('play')}
          </Button>
          <Button
            variant="secondary"
            size="lg"
            onClick={handleWatchlist}
          >
            <List className={cn('mr-2 h-5 w-5', inList && 'fill-current')} />
            {inList ? t('removeFromList') : t('addToList')}
          </Button>
          <Button
            variant="ghost"
            size="lg"
            onClick={() => router.push(`/movie/${movie.id}`)}
            aria-label={t('moreInfo')}
          >
            <Info className="mr-2 h-5 w-5" />
            {t('moreInfo')}
          </Button>
        </div>
      </div>
    </div>
  );
}

function HeroBannerSkeleton() {
  return (
    <div className="relative h-[80vh] min-h-[500px] w-full">
      <Skeleton variant="custom" className="h-full w-full" />
      <div className="absolute bottom-24 left-4 sm:left-8 lg:left-12 space-y-4 max-w-lg">
        <Skeleton variant="text" className="h-4 w-48" />
        <Skeleton variant="text" className="h-10 w-80" />
        <Skeleton variant="text" className="h-16 w-full" />
        <div className="flex gap-3">
          <Skeleton variant="custom" className="h-12 w-28 rounded-md" />
          <Skeleton variant="custom" className="h-12 w-36 rounded-md" />
        </div>
      </div>
    </div>
  );
}

export default function HeroBannerSection() {
  const { data: movie, isLoading } = useFeaturedMovie();

  if (isLoading) return <HeroBannerSkeleton />;
  if (!movie) return null;

  return <HeroBanner movie={movie} />;
}
