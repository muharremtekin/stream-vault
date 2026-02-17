'use client';

import { useTranslations } from 'next-intl';

import { AutocompleteItem } from '@/components/search/autocomplete-item';
import { Skeleton } from '@/components/ui/skeleton';
import type { AutocompleteSuggestion } from '@/lib/types/search';

const SKELETON_COUNT = 3;

interface AutocompleteDropdownProps {
  suggestions: AutocompleteSuggestion[];
  isLoading: boolean;
  query: string;
  onSelect: () => void;
}

export function AutocompleteDropdown({
  suggestions,
  isLoading,
  query,
  onSelect,
}: AutocompleteDropdownProps) {
  const t = useTranslations('search');

  if (!isLoading && suggestions.length === 0) {
    return null;
  }

  return (
    <div
      className="absolute right-0 top-full mt-1 w-72 overflow-hidden rounded-md border border-border bg-card shadow-lg sm:w-80"
      role="listbox"
      aria-label={t('ariaLabel')}
    >
      {isLoading ? (
        <div className="space-y-1 p-2">
          {Array.from({ length: SKELETON_COUNT }).map((_, i) => (
            <Skeleton
              key={`autocomplete-skeleton-${String(i)}`}
              variant="custom"
              className="h-9 w-full rounded"
            />
          ))}
        </div>
      ) : (
        <div className="py-1">
          {suggestions.map((suggestion) => (
            <AutocompleteItem
              key={suggestion.contentId}
              suggestion={suggestion}
              query={query}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}
