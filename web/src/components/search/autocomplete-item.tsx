'use client';

import { useRouter } from 'next/navigation';

import { Film, Tv } from 'lucide-react';

import { SearchHighlight } from '@/components/search/search-highlight';
import type { AutocompleteSuggestion } from '@/lib/types/search';

interface AutocompleteItemProps {
  suggestion: AutocompleteSuggestion;
  query: string;
  onSelect: () => void;
}

export function AutocompleteItem({ suggestion, query, onSelect }: AutocompleteItemProps) {
  const router = useRouter();

  const isSeries = suggestion.contentType.toLowerCase() === 'series';
  const path = isSeries
    ? `/series/${suggestion.contentId}`
    : `/movie/${suggestion.contentId}`;

  const handleClick = () => {
    onSelect();
    router.push(path);
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      className="flex w-full items-center gap-3 px-3 py-2 text-left transition-colors hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
    >
      {isSeries ? (
        <Tv className="h-4 w-4 shrink-0 text-muted-foreground" />
      ) : (
        <Film className="h-4 w-4 shrink-0 text-muted-foreground" />
      )}
      <span className="min-w-0 flex-1 truncate text-sm text-foreground">
        <SearchHighlight text={suggestion.title} query={query} />
      </span>
      {suggestion.releaseYear > 0 && (
        <span className="shrink-0 text-xs text-muted-foreground">
          {suggestion.releaseYear}
        </span>
      )}
    </button>
  );
}
