import { apiClient } from '@/lib/api/client';
import type { PaginatedResponse } from '@/lib/types/common';
import type {
  Movie,
  Series,
  Genre,
  ContentSummary,
  CatalogParams,
} from '@/lib/types/catalog';
import type { MovieFormData, SeriesFormData } from '@/lib/validations/content';

export async function getMovies(
  params?: CatalogParams
): Promise<PaginatedResponse<Movie>> {
  const response = await apiClient.get<PaginatedResponse<Movie>>(
    '/api/catalog/movies',
    { params }
  );
  return response.data;
}

export async function getMovie(id: string): Promise<Movie> {
  const response = await apiClient.get<Movie>(`/api/catalog/movies/${id}`);
  return response.data;
}

export async function getSeries(
  params?: CatalogParams
): Promise<PaginatedResponse<Series>> {
  const response = await apiClient.get<PaginatedResponse<Series>>(
    '/api/catalog/series',
    { params }
  );
  return response.data;
}

export async function getSeriesById(id: string): Promise<Series> {
  const response = await apiClient.get<Series>(`/api/catalog/series/${id}`);
  return response.data;
}

export async function getGenres(): Promise<Genre[]> {
  const response = await apiClient.get<Genre[]>('/api/catalog/genres');
  return response.data;
}

export async function getGenreContent(
  slug: string,
  params?: { page?: number; pageSize?: number }
): Promise<PaginatedResponse<ContentSummary>> {
  const response = await apiClient.get<PaginatedResponse<ContentSummary>>(
    `/api/catalog/genres/${slug}/content`,
    { params }
  );
  return response.data;
}

export async function createMovie(
  data: MovieFormData
): Promise<{ id: string }> {
  const response = await apiClient.post<{ id: string }>(
    '/api/catalog/movies',
    data
  );
  return response.data;
}

export async function createSeries(
  data: SeriesFormData
): Promise<{ id: string }> {
  const response = await apiClient.post<{ id: string }>(
    '/api/catalog/series',
    data
  );
  return response.data;
}

export interface AddEpisodeRequest {
  episodeNumber: number;
  title: string;
  description: string;
  durationMinutes: number;
  thumbnailUrl: string;
}

export async function addEpisode(
  seriesId: string,
  seasonNumber: number,
  data: AddEpisodeRequest
): Promise<void> {
  await apiClient.post(
    `/api/catalog/series/${seriesId}/seasons/${seasonNumber}/episodes`,
    data
  );
}
