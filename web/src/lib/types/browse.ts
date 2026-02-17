import type { ContentType } from '@/lib/types/common';

/**
 * Unified card item type — adapter for RecommendedItem, TrendingItem,
 * Movie, Series, and WatchProgress (cross-referenced with catalog data).
 */
export interface ContentCardItem {
  contentId: string;
  title: string;
  thumbnailUrl?: string;
  releaseYear?: number;
  averageRating?: number;
  genres?: string[];
  contentType?: ContentType | string;
  durationMinutes?: number;
  totalSeasons?: number;
  /** Filled for continue-watching rows (0–100) */
  percentage?: number;
  /** Filled for trending rows */
  rank?: number;
}
