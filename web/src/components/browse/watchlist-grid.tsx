'use client';

import { ListX } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

import { Skeleton } from '@/components/ui/skeleton';
import WatchlistItemCard from '@/components/browse/watchlist-item-card';
import { useWatchlist } from '@/lib/hooks/use-watchlist';
import { cn } from '@/lib/utils/cn';

const SKELETON_COUNT = 8;

function WatchlistSkeleton() {
  return (
    <div className="flex flex-wrap gap-2 md:gap-3">
      {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
        <div
          key={`skeleton-${i}`}
          className="shrink-0 w-32 sm:w-36 md:w-44"
        >
          <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
        </div>
      ))}
    </div>
  );
}

interface EmptyStateProps {
  title: string;
  description: string;
  ctaLabel: string;
}

function EmptyState({ title, description, ctaLabel }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <ListX className="mb-6 h-16 w-16 text-muted-foreground" />
      <h2 className="mb-2 text-xl font-semibold text-foreground">{title}</h2>
      <p className="mb-8 max-w-sm text-sm text-muted-foreground">{description}</p>
      <Link
        href="/browse"
        className={cn(
          'rounded-md bg-primary px-6 py-2.5 text-sm font-semibold text-primary-foreground',
          'transition-colors hover:bg-accent',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
        )}
      >
        {ctaLabel}
      </Link>
    </div>
  );
}

export default function WatchlistGrid() {
  const t = useTranslations('browse');
  const { items, isLoading } = useWatchlist();

  if (isLoading) {
    return <WatchlistSkeleton />;
  }

  if (items.length === 0) {
    return (
      <EmptyState
        title={t('emptyList')}
        description={t('emptyListDescription')}
        ctaLabel={t('browseContent')}
      />
    );
  }

  return (
    <div className="flex flex-wrap gap-2 md:gap-3">
      {items.map((item) => (
        <WatchlistItemCard key={item.id} item={item} />
      ))}
    </div>
  );
}
