import { cn } from '@/lib/utils/cn';

interface TrendingBadgeProps {
  rank: number;
  className?: string;
}

export default function TrendingBadge({ rank, className }: TrendingBadgeProps) {
  return (
    <div
      aria-label={`#${rank} trending`}
      className={cn(
        'absolute bottom-0 left-0 select-none pointer-events-none',
        'text-7xl font-black leading-none',
        'text-background drop-shadow-[0_0_4px_rgba(255,255,255,0.6)]',
        '[-webkit-text-stroke:2px_rgba(255,255,255,0.4)]',
        className
      )}
    >
      {rank}
    </div>
  );
}
