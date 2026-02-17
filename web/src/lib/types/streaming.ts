import type { VideoStatus } from '@/lib/types/common';

export interface QualityInfo {
  label: string;
  width: number;
  height: number;
  bitrateKbps: number;
  segmentCount: number;
}

export interface StreamingInfo {
  videoStatus: VideoStatus;
  durationSeconds: number;
  availableQualities: QualityInfo[];
  manifestUrl: string;
  thumbnailUrl: string | null;
  posterUrl: string | null;
  encodedAt: string | null;
}

/** Watch progress — Go streaming service returns snake_case JSON */
export interface WatchProgress {
  user_id: string;
  content_id: string;
  position_seconds: number;
  duration_seconds: number;
  percentage: number;
  updated_at: number;
}

export interface SaveProgressRequest {
  position_seconds: number;
  duration_seconds: number;
}

export interface ContinueWatchingResponse {
  items: WatchProgress[];
}

export interface NextEpisodeInfo {
  id: string;
  episodeNumber: number;
  seasonNumber: number;
  title: string;
  thumbnailUrl?: string | null;
}
