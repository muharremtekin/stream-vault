'use client';

import Image from 'next/image';
import { useRouter } from 'next/navigation';

import { useTranslations } from 'next-intl';

import { Skeleton } from '@/components/ui/skeleton';
import { useContent } from '@/lib/hooks/use-content';
import { cn } from '@/lib/utils/cn';
import type { Movie } from '@/lib/types/catalog';

interface ContentInfoProps {
  id: string;
  contentType: 'movie' | 'series';
  className?: string;
}

function isMovie(data: unknown): data is Movie {
  return typeof data === 'object' && data !== null && 'director' in data;
}

export default function ContentInfo({ id, contentType, className }: ContentInfoProps) {
  const router = useRouter();
  const tb = useTranslations('browse');
  const tc = useTranslations('content');
  const { data, isLoading } = useContent(id, contentType);

  if (isLoading || !data) {
    return (
      <div className={cn('space-y-6', className)}>
        <Skeleton variant="text" className="h-4 w-full" />
        <Skeleton variant="text" className="h-4 w-5/6" />
        <Skeleton variant="text" className="h-4 w-4/6" />
        <div className="grid grid-cols-2 gap-4 pt-4">
          <Skeleton variant="text" className="h-4 w-32" />
          <Skeleton variant="text" className="h-4 w-32" />
        </div>
      </div>
    );
  }

  const movie = isMovie(data) ? data : null;
  const credit = movie ? movie.director : ('creator' in data ? data.creator : '');
  const creditLabel = movie ? tb('director') : tc('creator');

  const handleGenreClick = (genre: string) => {
    const slug = genre.toLowerCase().replace(/\s+/g, '-');
    router.push(`/genre/${slug}`);
  };

  return (
    <div className={cn('space-y-6', className)}>
      {/* About section title */}
      <h2 className="text-xl font-semibold text-foreground">{tc('about')}</h2>

      {/* Description */}
      <p className="text-muted-foreground leading-relaxed">{data.description}</p>

      {/* Director / Creator */}
      {credit && (
        <div>
          <span className="text-sm font-semibold text-foreground">{creditLabel}: </span>
          <span className="text-sm text-muted-foreground">{credit}</span>
        </div>
      )}

      {/* Cast */}
      {data.cast?.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-foreground">{tb('cast')}</h3>
          <div className="flex flex-wrap gap-3">
            {data.cast.slice(0, 6).map((member) => (
              <div key={member.name} className="flex items-center gap-2">
                <div className="relative h-8 w-8 shrink-0 overflow-hidden rounded-full bg-muted">
                  {member.photoUrl ? (
                    <Image
                      src={member.photoUrl}
                      alt={member.name}
                      fill
                      sizes="32px"
                      className="object-cover"
                    />
                  ) : null}
                </div>
                <div>
                  <p className="text-xs font-medium text-foreground">{member.name}</p>
                  <p className="text-xs text-muted-foreground">{member.role}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Genres */}
      {data.genres?.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-foreground">{tb('genres')}</h3>
          <div className="flex flex-wrap gap-2">
            {data.genres.map((genre) => (
              <button
                key={genre}
                type="button"
                onClick={() => handleGenreClick(genre)}
                className={cn(
                  'rounded-full px-3 py-1 text-xs',
                  'bg-white/10 text-foreground',
                  'hover:bg-white/20 transition-colors',
                  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
                )}
              >
                {genre}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Tags */}
      {data.tags?.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {data.tags.map((tag) => (
            <span
              key={tag}
              className="rounded bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {tag}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
