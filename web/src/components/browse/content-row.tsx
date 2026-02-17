'use client';

import { useRef, useState, useCallback } from 'react';

import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils/cn';

interface ContentRowProps {
  title: string;
  children?: React.ReactNode;
  isLoading?: boolean;
  className?: string;
}

const CARD_SCROLL_AMOUNT = 600;
const SKELETON_CARD_COUNT = 6;

export default function ContentRow({
  title,
  children,
  isLoading = false,
  className,
}: ContentRowProps) {
  const t = useTranslations('browse');
  const scrollRef = useRef<HTMLDivElement>(null);
  const [showLeft, setShowLeft] = useState(false);
  const [showRight, setShowRight] = useState(true);

  const updateArrows = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setShowLeft(el.scrollLeft > 8);
    setShowRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 8);
  }, []);

  const scrollLeft = () => {
    scrollRef.current?.scrollBy({ left: -CARD_SCROLL_AMOUNT, behavior: 'smooth' });
  };

  const scrollRight = () => {
    scrollRef.current?.scrollBy({ left: CARD_SCROLL_AMOUNT, behavior: 'smooth' });
  };

  return (
    <section className={cn('group/row px-4 sm:px-8 lg:px-12', className)}>
      <h2 className="mb-3 text-lg font-semibold text-foreground sm:text-xl">
        {title}
      </h2>

      <div className="relative">
        {/* Left arrow */}
        {showLeft && (
          <button
            type="button"
            onClick={scrollLeft}
            aria-label={t('scrollLeft')}
            className={cn(
              'absolute left-0 top-0 z-20 h-full w-10',
              'flex items-center justify-center',
              'bg-gradient-to-r from-background/80 to-transparent',
              'opacity-0 group-hover/row:opacity-100 transition-opacity'
            )}
          >
            <ChevronLeft className="h-6 w-6 text-foreground" />
          </button>
        )}

        {/* Scroll container */}
        <div
          ref={scrollRef}
          onScroll={updateArrows}
          className="flex gap-2 overflow-x-auto scroll-smooth pb-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        >
          {isLoading
            ? Array.from({ length: SKELETON_CARD_COUNT }).map((_, i) => (
                // eslint-disable-next-line react/no-array-index-key
                <div key={`skeleton-${i}`} className="shrink-0 w-32 sm:w-36 md:w-44">
                  <Skeleton variant="card" className="aspect-[2/3] w-full rounded-md" />
                </div>
              ))
            : children}
        </div>

        {/* Right arrow */}
        {showRight && (
          <button
            type="button"
            onClick={scrollRight}
            aria-label={t('scrollRight')}
            className={cn(
              'absolute right-0 top-0 z-20 h-full w-10',
              'flex items-center justify-center',
              'bg-gradient-to-l from-background/80 to-transparent',
              'opacity-0 group-hover/row:opacity-100 transition-opacity'
            )}
          >
            <ChevronRight className="h-6 w-6 text-foreground" />
          </button>
        )}
      </div>
    </section>
  );
}
