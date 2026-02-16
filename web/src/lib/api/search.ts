import { apiClient } from '@/lib/api/client';
import type {
  SearchResponse,
  SearchParams,
  AutocompleteResponse,
  TrendingResponse,
} from '@/lib/types/search';

export async function search(params: SearchParams): Promise<SearchResponse> {
  const response = await apiClient.get<SearchResponse>('/api/search', {
    params,
  });
  return response.data;
}

export async function autocomplete(
  query: string,
  limit?: number
): Promise<AutocompleteResponse> {
  const response = await apiClient.get<AutocompleteResponse>(
    '/api/search/autocomplete',
    { params: { q: query, ...(limit ? { limit } : {}) } }
  );
  return response.data;
}

export async function getTrending(
  window?: string,
  limit?: number
): Promise<TrendingResponse> {
  const response = await apiClient.get<TrendingResponse>(
    '/api/search/trending',
    { params: { ...(window ? { window } : {}), ...(limit ? { limit } : {}) } }
  );
  return response.data;
}
