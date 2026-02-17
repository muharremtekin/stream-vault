'use client';

import { useState, useEffect, useRef, useCallback } from 'react';

import { useRouter } from 'next/navigation';
import { Search, X } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { AutocompleteDropdown } from '@/components/search/autocomplete-dropdown';
import { cn } from '@/lib/utils/cn';
import { useAutocomplete } from '@/lib/hooks/use-search';
import { SEARCH_MIN_CHARS } from '@/lib/utils/constants';

export function SearchBar() {
  const t = useTranslations('search');
  const router = useRouter();
  const [isExpanded, setIsExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const { query, setQuery, suggestions, isLoading } = useAutocomplete();

  const collapse = useCallback(() => {
    setIsExpanded(false);
    setQuery('');
  }, [setQuery]);

  useEffect(() => {
    if (isExpanded) {
      inputRef.current?.focus();
    }
  }, [isExpanded]);

  useEffect(() => {
    if (!isExpanded) return;

    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        collapse();
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isExpanded, collapse]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && query.length >= SEARCH_MIN_CHARS) {
      e.preventDefault();
      collapse();
      router.push(`/search?q=${encodeURIComponent(query)}`);
    }
    if (e.key === 'Escape') {
      collapse();
    }
  };

  const handleToggle = () => {
    if (isExpanded) {
      collapse();
    } else {
      setIsExpanded(true);
    }
  };

  const handleAutocompleteSelect = () => {
    collapse();
  };

  const showDropdown =
    isExpanded && query.length >= SEARCH_MIN_CHARS && (isLoading || suggestions.length > 0);

  return (
    <div ref={containerRef} className="relative flex items-center">
      <div
        className={cn(
          'flex items-center overflow-hidden rounded-full transition-all duration-300',
          isExpanded
            ? 'border border-border bg-card'
            : 'border border-transparent'
        )}
      >
        <button
          type="button"
          onClick={handleToggle}
          aria-label={t('ariaLabel')}
          className="shrink-0 rounded-full p-2 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          <Search className="h-5 w-5" />
        </button>

        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t('placeholder')}
          aria-label={t('ariaLabel')}
          className={cn(
            'bg-transparent text-sm text-foreground placeholder:text-muted-foreground transition-all duration-300 focus:outline-none',
            isExpanded ? 'w-36 pr-2 opacity-100 sm:w-52' : 'w-0 opacity-0'
          )}
        />

        {isExpanded && query.length > 0 && (
          <button
            type="button"
            onClick={() => setQuery('')}
            aria-label={t('close')}
            className="shrink-0 pr-2 text-muted-foreground transition-colors hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {showDropdown && (
        <AutocompleteDropdown
          suggestions={suggestions}
          isLoading={isLoading}
          query={query}
          onSelect={handleAutocompleteSelect}
        />
      )}
    </div>
  );
}
