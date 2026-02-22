import { cn } from '@/lib/utils/cn';

interface MaturityBadgeProps {
  rating: string;
  size?: 'sm' | 'md';
  className?: string;
}

const MATURITY_COLORS: Record<string, string> = {
  G: 'border-green-500 text-green-400',
  PG: 'border-yellow-500 text-yellow-400',
  PG13: 'border-orange-500 text-orange-400',
  R: 'border-red-500 text-red-400',
  NC17: 'border-red-700 text-red-500',
};

/** User-friendly display labels for enum-style rating values */
const MATURITY_LABELS: Record<string, string> = {
  PG13: 'PG-13',
  NC17: 'NC-17',
};

const SIZES = {
  sm: 'px-1.5 py-0.5 text-[10px]',
  md: 'px-2 py-0.5 text-xs',
} as const;

export default function MaturityBadge({
  rating,
  size = 'md',
  className,
}: MaturityBadgeProps) {
  const colorClass = MATURITY_COLORS[rating] ?? 'border-muted-foreground text-muted-foreground';

  return (
    <span
      className={cn(
        'inline-block border font-semibold rounded',
        SIZES[size],
        colorClass,
        className
      )}
    >
      {MATURITY_LABELS[rating] ?? rating}
    </span>
  );
}
