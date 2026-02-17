'use client';

import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

interface SearchPaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function getPageNumbers(current: number, total: number): (number | 'ellipsis')[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | 'ellipsis')[] = [1];

  if (current > 3) pages.push('ellipsis');

  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  if (current < total - 2) pages.push('ellipsis');

  pages.push(total);
  return pages;
}

export function SearchPagination({ currentPage, totalPages, onPageChange }: SearchPaginationProps) {
  const t = useTranslations('search');

  if (totalPages <= 1) return null;

  const pages = getPageNumbers(currentPage, totalPages);

  return (
    <nav aria-label={t('page', { current: currentPage, total: totalPages })} className="flex items-center justify-center gap-1">
      <button
        type="button"
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage <= 1}
        aria-label={t('previousPage')}
        className={cn(
          'rounded-md p-2 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          currentPage <= 1
            ? 'cursor-not-allowed text-muted-foreground/30'
            : 'text-muted-foreground hover:text-foreground'
        )}
      >
        <ChevronLeft className="h-4 w-4" />
      </button>

      {pages.map((page, i) => {
        if (page === 'ellipsis') {
          return (
            <span key={`ellipsis-${String(i)}`} className="px-2 text-sm text-muted-foreground">
              ...
            </span>
          );
        }

        return (
          <button
            key={page}
            type="button"
            onClick={() => onPageChange(page)}
            className={cn(
              'min-w-8 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
              page === currentPage
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            {page}
          </button>
        );
      })}

      <button
        type="button"
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage >= totalPages}
        aria-label={t('nextPage')}
        className={cn(
          'rounded-md p-2 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          currentPage >= totalPages
            ? 'cursor-not-allowed text-muted-foreground/30'
            : 'text-muted-foreground hover:text-foreground'
        )}
      >
        <ChevronRight className="h-4 w-4" />
      </button>
    </nav>
  );
}
