import { apiClient } from '@/lib/api/client';
import type {
  RecommendationsResponse,
  SimilarResponse,
  HomePageResponse,
  FeedbackRequest,
} from '@/lib/types/recommendation';

export async function getRecommendations(
  limit?: number
): Promise<RecommendationsResponse> {
  const response = await apiClient.get<RecommendationsResponse>(
    '/api/recommendations',
    { params: limit ? { limit } : undefined }
  );
  return response.data;
}

export async function getSimilar(
  contentId: string,
  limit?: number
): Promise<SimilarResponse> {
  const response = await apiClient.get<SimilarResponse>(
    `/api/recommendations/similar/${contentId}`,
    { params: limit ? { limit } : undefined }
  );
  return response.data;
}

export async function getHomeSections(): Promise<HomePageResponse> {
  const response =
    await apiClient.get<HomePageResponse>('/api/recommendations/home');
  return response.data;
}

export async function sendFeedback(data: FeedbackRequest): Promise<void> {
  await apiClient.post('/api/recommendations/feedback', data);
}
