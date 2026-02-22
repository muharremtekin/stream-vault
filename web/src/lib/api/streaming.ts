import axios from 'axios';

import { apiClient } from '@/lib/api/client';
import { API_URL } from '@/lib/utils/constants';
import type {
  StreamingInfo,
  WatchProgress,
  SaveProgressRequest,
  ContinueWatchingResponse,
} from '@/lib/types/streaming';

export async function getStreamingInfo(
  contentId: string
): Promise<StreamingInfo> {
  const response = await apiClient.get<StreamingInfo>(
    `/api/stream/${contentId}/info`
  );
  return response.data;
}

export async function getProgress(contentId: string): Promise<WatchProgress> {
  const response = await apiClient.get<WatchProgress>(
    `/api/stream/${contentId}/progress`
  );
  return response.data;
}

export async function saveProgress(
  contentId: string,
  data: SaveProgressRequest
): Promise<{ status: string }> {
  const response = await apiClient.post<{ status: string }>(
    `/api/stream/${contentId}/progress`,
    data
  );
  return response.data;
}

export async function getContinueWatching(
  limit?: number
): Promise<ContinueWatchingResponse> {
  const response = await apiClient.get<ContinueWatchingResponse>(
    '/api/stream/continue-watching',
    { params: limit ? { limit } : undefined }
  );
  return response.data;
}

export interface UploadResponse {
  job_id: string;
  content_id: string;
  status: string;
}

/**
 * Upload a video file for encoding (admin-only).
 * Uses a separate axios instance with extended timeout for large files.
 */
export async function uploadVideo(
  formData: FormData,
  onUploadProgress?: (percentage: number) => void
): Promise<UploadResponse> {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { useAuthStore } = require('@/lib/stores/auth-store') as {
    useAuthStore: {
      getState: () => { accessToken: string | null };
    };
  };

  const { accessToken } = useAuthStore.getState();

  const response = await axios.post<UploadResponse>(`${API_URL}/api/stream/upload`, formData, {
    timeout: 5 * 60_000, // 5 minutes for large uploads
    headers: {
      'Content-Type': 'multipart/form-data',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    onUploadProgress: (event) => {
      if (onUploadProgress && event.total) {
        const percentage = Math.round((event.loaded * 100) / event.total);
        onUploadProgress(percentage);
      }
    },
  });
  return response.data;
}
