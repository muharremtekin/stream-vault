'use client';

import { useTranslations } from 'next-intl';

import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils/cn';
import type { Facets, SearchParams } from '@/lib/types/search';

interface SearchFiltersProps {
  facets: Facets | undefined;
  currentParams: SearchParams;
  onFilterChange: (params: Partial<SearchParams>) => void;
}

const YEAR_RANGES = [
  { key: 'year2020Plus', from: 2020, to: undefined },
  { key: 'year2015to2019', from: 2015, to: 2019 },
  { key: 'year2010to2014', from: 2010, to: 2014 },
  { key: 'yearOlder', from: undefined, to: 2009 },
] as const;

const RATING_OPTIONS = [7, 8, 9] as const;

const SORT_OPTIONS = [
  { key: 'sortRelevance', value: 'relevance' },
  { key: 'sortRating', value: 'rating' },
  { key: 'sortYear', value: 'year' },
  { key: 'sortTitle', value: 'title' },
] as const;

export function SearchFilters({ facets, currentParams, onFilterChange }: SearchFiltersProps) {
  const t = useTranslations('search');

  const activeGenres = currentParams.genres?.split(',').filter(Boolean) ?? [];
  const activeType = currentParams.content_type ?? '';
  const activeSort = currentParams.sort ?? 'relevance';
  const activeYearFrom = currentParams.year_from;
  const activeYearTo = currentParams.year_to;
  const activeRating = currentParams.min_rating;

  const hasActiveFilters =
    activeGenres.length > 0 ||
    activeType !== '' ||
    activeYearFrom !== undefined ||
    activeRating !== undefined ||
    activeSort !== 'relevance';

  const handleGenreToggle = (genre: string) => {
    const updated = activeGenres.includes(genre)
      ? activeGenres.filter((g) => g !== genre)
      : [...activeGenres, genre];
    onFilterChange({ genres: updated.length > 0 ? updated.join(',') : undefined });
  };

  const handleYearRange = (from: number | undefined, to: number | undefined) => {
    const isSame = activeYearFrom === from && activeYearTo === to;
    onFilterChange({
      year_from: isSame ? undefined : from,
      year_to: isSame ? undefined : to,
    });
  };

  const handleRating = (value: number) => {
    onFilterChange({ min_rating: activeRating === value ? undefined : value });
  };

  const handleType = (value: string) => {
    onFilterChange({ content_type: activeType === value ? undefined : value });
  };

  const handleSort = (value: string) => {
    onFilterChange({ sort: value === 'relevance' ? undefined : value });
  };

  const handleClearAll = () => {
    onFilterChange({
      genres: undefined,
      content_type: undefined,
      year_from: undefined,
      year_to: undefined,
      min_rating: undefined,
      sort: undefined,
      order: undefined,
    });
  };

  return (
    <aside className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-foreground">{t('filters')}</h2>
        {hasActiveFilters && (
          <button
            type="button"
            onClick={handleClearAll}
            className="text-xs text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
          >
            {t('clearFilters')}
          </button>
        )}
      </div>

      {/* Active filter badges */}
      {hasActiveFilters && (
        <div className="flex flex-wrap gap-1.5">
          {activeGenres.map((genre) => (
            <Badge key={genre} size="sm" isRemovable onRemove={() => handleGenreToggle(genre)}>
              {genre}
            </Badge>
          ))}
          {activeType && (
            <Badge size="sm" isRemovable onRemove={() => handleType(activeType)}>
              {activeType === 'Movie' ? t('film') : t('series')}
            </Badge>
          )}
          {activeRating !== undefined && (
            <Badge size="sm" isRemovable onRemove={() => handleRating(activeRating)}>
              {t('ratingPlus', { value: activeRating })}
            </Badge>
          )}
        </div>
      )}

      {/* Content Type */}
      <FilterSection title={t('contentType')}>
        <div className="flex flex-wrap gap-1.5">
          {['Movie', 'Series'].map((type) => (
            <FilterChip
              key={type}
              label={type === 'Movie' ? t('film') : t('series')}
              isActive={activeType === type}
              onClick={() => handleType(type)}
            />
          ))}
        </div>
      </FilterSection>

      {/* Genres */}
      {facets?.genreFacets && facets.genreFacets.length > 0 && (
        <FilterSection title={t('genres')}>
          <div className="flex flex-wrap gap-1.5">
            {facets.genreFacets.map((facet) => (
              <FilterChip
                key={facet.key}
                label={`${facet.key} (${facet.docCount})`}
                isActive={activeGenres.includes(facet.key)}
                onClick={() => handleGenreToggle(facet.key)}
              />
            ))}
          </div>
        </FilterSection>
      )}

      {/* Year Range */}
      <FilterSection title={t('yearRange')}>
        <div className="flex flex-wrap gap-1.5">
          {YEAR_RANGES.map((range) => (
            <FilterChip
              key={range.key}
              label={t(range.key)}
              isActive={activeYearFrom === range.from && activeYearTo === range.to}
              onClick={() => handleYearRange(range.from, range.to)}
            />
          ))}
        </div>
      </FilterSection>

      {/* Min Rating */}
      <FilterSection title={t('rating')}>
        <div className="flex flex-wrap gap-1.5">
          {RATING_OPTIONS.map((value) => (
            <FilterChip
              key={value}
              label={t('ratingPlus', { value })}
              isActive={activeRating === value}
              onClick={() => handleRating(value)}
            />
          ))}
        </div>
      </FilterSection>

      {/* Sort */}
      <FilterSection title={t('sortBy')}>
        <div className="flex flex-wrap gap-1.5">
          {SORT_OPTIONS.map((option) => (
            <FilterChip
              key={option.value}
              label={t(option.key)}
              isActive={activeSort === option.value}
              onClick={() => handleSort(option.value)}
            />
          ))}
        </div>
      </FilterSection>
    </aside>
  );
}

function FilterSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="mb-2 text-xs font-medium text-muted-foreground uppercase tracking-wider">{title}</h3>
      {children}
    </div>
  );
}

interface FilterChipProps {
  label: string;
  isActive: boolean;
  onClick: () => void;
}

function FilterChip({ label, isActive, onClick }: FilterChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-full border px-3 py-1 text-xs font-medium transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        isActive
          ? 'border-primary bg-primary/20 text-primary'
          : 'border-border text-muted-foreground hover:border-foreground/30 hover:text-foreground'
      )}
    >
      {label}
    </button>
  );
}
