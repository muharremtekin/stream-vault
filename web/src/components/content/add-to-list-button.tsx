'use client';

import { Check, Plus, Loader2 } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { useWatchlist } from '@/lib/hooks/use-watchlist';
import { cn } from '@/lib/utils/cn';
import type { ContentType } from '@/lib/types/common';

interface AddToListButtonProps {
  contentId: string;
  contentType: ContentType;
  variant?: 'icon' | 'button';
  className?: string;
}

export default function AddToListButton({
  contentId,
  contentType,
  variant = 'button',
  className,
}: AddToListButtonProps) {
  const tb = useTranslations('browse');
  const ta = useTranslations('api');
  const { isInWatchlist, getWatchlistItemId, add, remove, isAdding, isRemoving } =
    useWatchlist();

  const inList = isInWatchlist(contentId);
  const isPending = isAdding || isRemoving;

  const handleToggle = async () => {
    if (isPending) return;
    if (inList) {
      const itemId = getWatchlistItemId(contentId);
      if (!itemId) return;
      await remove(itemId);
      toast.success(ta('removedFromList'));
    } else {
      await add({ contentId, contentType });
      toast.success(ta('addedToList'));
    }
  };

  if (variant === 'icon') {
    return (
      <button
        type="button"
        onClick={handleToggle}
        disabled={isPending}
        aria-label={inList ? tb('removeFromList') : tb('addToList')}
        className={cn(
          'flex h-9 w-9 items-center justify-center rounded-full border-2',
          'border-white/70 text-white hover:border-white transition-colors',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          'disabled:opacity-50',
          className
        )}
      >
        {isPending ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : inList ? (
          <Check className="h-4 w-4" />
        ) : (
          <Plus className="h-4 w-4" />
        )}
      </button>
    );
  }

  return (
    <Button
      variant="secondary"
      onClick={handleToggle}
      disabled={isPending}
      className={cn('gap-2', className)}
    >
      {isPending ? (
        <Loader2 className="h-4 w-4 animate-spin" />
      ) : inList ? (
        <Check className="h-4 w-4" />
      ) : (
        <Plus className="h-4 w-4" />
      )}
      {inList ? tb('removeFromList') : tb('addToList')}
    </Button>
  );
}
