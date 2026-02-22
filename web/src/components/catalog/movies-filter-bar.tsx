'use client';

import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

import type { Genre, CatalogParams } from '@/lib/types/catalog';

interface MoviesFilterBarProps {
  genres: Genre[];
  activeGenre: string | undefined;
  activeSort: string | undefined;
  onFilterChange: (updates: Partial<CatalogParams>) => void;
  isLoading: boolean;
}

const SORT_OPTIONS = [
  { key: 'sortNewest', value: 'newest' },
  { key: 'sortRating', value: 'rating' },
  { key: 'sortYear', value: 'year' },
  { key: 'sortTitle', value: 'title' },
] as const;

export function MoviesFilterBar({
  genres,
  activeGenre,
  activeSort,
  onFilterChange,
  isLoading,
}: MoviesFilterBarProps) {
  const t = useTranslations('pages.movies');

  const hasActiveFilters = !!activeGenre || (!!activeSort && activeSort !== 'newest');

  const handleGenreChange = (slug: string | undefined) => {
    onFilterChange({ genre: slug });
  };

  const handleSortChange = (value: string) => {
    onFilterChange({ sort: value === 'newest' ? undefined : value });
  };

  return (
    <div className="mb-8 space-y-4">
      {/* Genre chips */}
      <div className="flex items-center gap-3">
        <div className="flex flex-wrap gap-2">
          <FilterChip
            label={t('allGenres')}
            isActive={!activeGenre}
            isDisabled={isLoading}
            onClick={() => handleGenreChange(undefined)}
          />
          {genres.map((genre) => (
            <FilterChip
              key={genre.id}
              label={genre.name}
              isActive={activeGenre === genre.slug}
              isDisabled={isLoading}
              onClick={() =>
                handleGenreChange(activeGenre === genre.slug ? undefined : genre.slug)
              }
            />
          ))}
        </div>
      </div>

      {/* Sort + Clear row */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          {t('sortBy')}
        </span>
        {SORT_OPTIONS.map((option) => (
          <FilterChip
            key={option.value}
            label={t(option.key)}
            isActive={(activeSort ?? 'newest') === option.value}
            isDisabled={isLoading}
            onClick={() => handleSortChange(option.value)}
          />
        ))}

        {hasActiveFilters && (
          <button
            type="button"
            onClick={() => onFilterChange({ genre: undefined, sort: undefined })}
            className="ml-2 rounded text-xs text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            {t('clearFilters')}
          </button>
        )}
      </div>
    </div>
  );
}

interface FilterChipProps {
  label: string;
  isActive: boolean;
  isDisabled: boolean;
  onClick: () => void;
}

function FilterChip({ label, isActive, isDisabled, onClick }: FilterChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isDisabled}
      className={cn(
        'rounded-full border px-3 py-1 text-xs font-medium transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        'disabled:pointer-events-none disabled:opacity-50',
        isActive
          ? 'border-primary bg-primary/20 text-primary'
          : 'border-border text-muted-foreground hover:border-foreground/30 hover:text-foreground'
      )}
    >
      {label}
    </button>
  );
}
