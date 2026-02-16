export interface EncodingOutput {
  quality: string;
  width: number;
  height: number;
  bitrateKbps: number;
  fileSizeBytes: number;
  segmentCount: number;
  playlistPath: string;
  storagePath: string;
}

export interface EncodingJob {
  jobId: string;
  contentId: string;
  status: string;
  progressPercentage: number;
  currentStep: string;
  outputs: EncodingOutput[];
  errorMessage?: string;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
}

export interface EncodingJobsResponse {
  jobs: EncodingJob[];
  totalCount: number;
}

export interface EncodingJobsParams {
  status?: string;
  content_id?: string;
  limit?: number;
  offset?: number;
}
