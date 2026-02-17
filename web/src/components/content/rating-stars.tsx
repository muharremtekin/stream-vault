import { Star } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

interface RatingStarsProps {
  rating: number;
  totalRatings?: number;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const SIZES = {
  sm: { star: 'h-3 w-3', text: 'text-xs' },
  md: { star: 'h-4 w-4', text: 'text-sm' },
  lg: { star: 'h-5 w-5', text: 'text-base' },
} as const;

export default function RatingStars({
  rating,
  totalRatings,
  size = 'md',
  className,
}: RatingStarsProps) {
  const t = useTranslations('content');
  const { star: starClass, text: textClass } = SIZES[size];

  // Convert 0–10 scale to 0–5 stars
  const stars = rating / 2;
  const fullStars = Math.floor(stars);
  const hasHalf = stars - fullStars >= 0.5;

  return (
    <div className={cn('flex items-center gap-1.5', className)}>
      <div className="flex items-center gap-0.5" aria-label={`${rating} out of 10`}>
        {Array.from({ length: 5 }).map((_, i) => {
          const isFilled = i < fullStars;
          const isHalf = !isFilled && i === fullStars && hasHalf;
          return (
            // eslint-disable-next-line react/no-array-index-key
            <Star
              key={`star-${i}`}
              className={cn(
                starClass,
                isFilled || isHalf
                  ? 'text-yellow-400 fill-yellow-400'
                  : 'text-muted-foreground'
              )}
              style={
                isHalf
                  ? {
                      clipPath: 'inset(0 50% 0 0)',
                      /* dynamic clip for half-star display */
                    }
                  : undefined
              }
            />
          );
        })}
      </div>
      <span className={cn('text-muted-foreground', textClass)}>
        {rating.toFixed(1)}
        {totalRatings != null && (
          <span className="ml-1">
            ({t('totalRatings', { count: totalRatings })})
          </span>
        )}
      </span>
    </div>
  );
}
