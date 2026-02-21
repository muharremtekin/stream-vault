'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import {
  getMovies,
  getSeries,
  getMovie,
  getSeriesById,
  createMovie,
  createSeries,
} from '@/lib/api/catalog';
import { getJobs, getJob } from '@/lib/api/encoding';

import type { CatalogParams } from '@/lib/types/catalog';
import type { EncodingJobsParams } from '@/lib/types/encoding';
import type { MovieFormData, SeriesFormData } from '@/lib/validations/content';

const STALE_2MIN = 2 * 60_000;

export function useAdminMovies(params?: CatalogParams) {
  return useQuery({
    queryKey: ['admin', 'movies', params],
    queryFn: () => getMovies(params),
    staleTime: STALE_2MIN,
  });
}

export function useAdminSeries(params?: CatalogParams) {
  return useQuery({
    queryKey: ['admin', 'series', params],
    queryFn: () => getSeries(params),
    staleTime: STALE_2MIN,
  });
}

export function useAdminMovie(id: string) {
  return useQuery({
    queryKey: ['admin', 'movie', id],
    queryFn: () => getMovie(id),
    staleTime: STALE_2MIN,
    enabled: !!id,
  });
}

export function useAdminSeriesDetail(id: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['admin', 'seriesDetail', id],
    queryFn: () => getSeriesById(id),
    staleTime: STALE_2MIN,
    enabled: (options?.enabled ?? true) && !!id,
    retry: options?.enabled === undefined ? undefined : 0,
  });
}

export function useAdminEncodingJobs(params?: EncodingJobsParams) {
  return useQuery({
    queryKey: ['admin', 'encoding', 'jobs', params],
    queryFn: () => getJobs(params),
    staleTime: STALE_2MIN,
  });
}

export function useAdminEncodingJob(id: string) {
  return useQuery({
    queryKey: ['admin', 'encoding', 'job', id],
    queryFn: () => getJob(id),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === 'processing' || status === 'pending') return 5_000;
      return false;
    },
  });
}

export function useCreateMovie() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: MovieFormData) => createMovie(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'movies'] });
    },
  });
}

export function useCreateSeries() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: SeriesFormData) => createSeries(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'series'] });
    },
  });
}
