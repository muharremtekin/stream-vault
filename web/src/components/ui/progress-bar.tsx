'use client';

import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const PROGRESS_BAR_VARIANTS = {
  primary: 'bg-primary',
  success: 'bg-success',
  warning: 'bg-warning',
} as const;

type ProgressBarVariant = keyof typeof PROGRESS_BAR_VARIANTS;

const PROGRESS_BAR_SIZES = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-3',
} as const;

type ProgressBarSize = keyof typeof PROGRESS_BAR_SIZES;

interface ProgressBarProps {
  value: number;
  variant?: ProgressBarVariant;
  size?: ProgressBarSize;
  showLabel?: boolean;
  label?: string;
  className?: string;
}

export function ProgressBar({
  value,
  variant = 'primary',
  size = 'md',
  showLabel = false,
  label,
  className,
}: ProgressBarProps) {
  const t = useTranslations('ui.progressBar');
  const clampedValue = Math.min(100, Math.max(0, value));

  return (
    <div className={cn('w-full', className)}>
      {showLabel && (
        <div className="mb-1 flex items-center justify-between text-xs text-muted-foreground">
          {label && <span>{label}</span>}
          <span className={cn(!label && 'ml-auto')}>{Math.round(clampedValue)}%</span>
        </div>
      )}

      <div
        className={cn('w-full overflow-hidden rounded-full bg-muted', PROGRESS_BAR_SIZES[size])}
        role="progressbar"
        aria-valuenow={clampedValue}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={t('progress', { value: Math.round(clampedValue) })}
      >
        <div
          className={cn(
            'h-full rounded-full transition-all duration-300 ease-out',
            PROGRESS_BAR_VARIANTS[variant]
          )}
          /* dynamic width — requires inline style */
          style={{ width: `${clampedValue}%` }}
        />
      </div>
    </div>
  );
}
