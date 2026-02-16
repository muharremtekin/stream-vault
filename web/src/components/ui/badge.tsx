'use client';

import { X } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const BADGE_VARIANTS = {
  default: 'bg-muted text-foreground',
  success: 'bg-success/20 text-success',
  warning: 'bg-warning/20 text-warning',
  error: 'bg-destructive/20 text-destructive',
  info: 'bg-info/20 text-info',
} as const;

type BadgeVariant = keyof typeof BADGE_VARIANTS;

const BADGE_SIZES = {
  sm: 'text-[10px] px-1.5 py-0.5',
  md: 'text-xs px-2.5 py-0.5',
} as const;

type BadgeSize = keyof typeof BADGE_SIZES;

interface BadgeProps {
  variant?: BadgeVariant;
  size?: BadgeSize;
  children: React.ReactNode;
  isRemovable?: boolean;
  onRemove?: () => void;
  className?: string;
}

export function Badge({
  variant = 'default',
  size = 'md',
  children,
  isRemovable = false,
  onRemove,
  className,
}: BadgeProps) {
  const t = useTranslations('ui.badge');

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full font-medium',
        BADGE_VARIANTS[variant],
        BADGE_SIZES[size],
        className
      )}
    >
      {children}
      {isRemovable && onRemove && (
        <button
          type="button"
          onClick={onRemove}
          aria-label={t('remove', { label: String(children) })}
          className="rounded-full p-0.5 transition-colors hover:bg-foreground/10 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </span>
  );
}
