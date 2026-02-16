'use client';

import { useState, useEffect, useRef } from 'react';

import { useQuery } from '@tanstack/react-query';

import { search, autocomplete } from '@/lib/api/search';
import { SEARCH_DEBOUNCE, SEARCH_MIN_CHARS } from '@/lib/utils/constants';
import type { SearchParams } from '@/lib/types/search';

export function useSearch(params: SearchParams) {
  return useQuery({
    queryKey: ['search', 'results', params],
    queryFn: () => search(params),
    enabled: !!params.q && params.q.length >= SEARCH_MIN_CHARS,
    staleTime: 30_000,
  });
}

export function useAutocomplete() {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);

    timerRef.current = setTimeout(() => {
      setDebouncedQuery(query);
    }, SEARCH_DEBOUNCE);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [query]);

  const autocompleteQuery = useQuery({
    queryKey: ['search', 'autocomplete', debouncedQuery],
    queryFn: () => autocomplete(debouncedQuery),
    enabled: debouncedQuery.length >= SEARCH_MIN_CHARS,
    staleTime: 60_000,
  });

  return {
    query,
    setQuery,
    suggestions: autocompleteQuery.data?.suggestions ?? [],
    isLoading: autocompleteQuery.isLoading,
  };
}
