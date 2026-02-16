import { apiClient } from '@/lib/api/client';
import type {
  EncodingJob,
  EncodingJobsResponse,
  EncodingJobsParams,
} from '@/lib/types/encoding';

export async function getJobs(
  params?: EncodingJobsParams
): Promise<EncodingJobsResponse> {
  const response = await apiClient.get<EncodingJobsResponse>(
    '/api/encoding/jobs',
    { params }
  );
  return response.data;
}

export async function getJob(id: string): Promise<EncodingJob> {
  const response = await apiClient.get<EncodingJob>(
    `/api/encoding/jobs/${id}`
  );
  return response.data;
}
