import type { ContentType } from '@/lib/types/common';

export interface CastMember {
  name: string;
  role: string;
  photoUrl: string | null;
}

export interface Movie {
  id: string;
  title: string;
  originalTitle: string | null;
  description: string;
  releaseYear: number;
  durationMinutes: number;
  durationFormatted: string;
  maturityRating: string;
  genres: string[];
  cast: CastMember[];
  director: string;
  thumbnailUrl: string;
  bannerUrl: string;
  trailerUrl: string | null;
  averageRating: number;
  ratingCount: number;
  status: string;
  videoStatus: string;
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

export interface Episode {
  episodeNumber: number;
  title: string;
  description: string;
  durationMinutes: number;
  durationFormatted: string;
  thumbnailUrl: string;
  videoStatus: string;
}

export interface Season {
  seasonNumber: number;
  title: string | null;
  releaseYear: number;
  episodes: Episode[];
  episodeCount: number;
}

export interface Series {
  id: string;
  title: string;
  originalTitle: string | null;
  description: string;
  releaseYear: number;
  maturityRating: string;
  genres: string[];
  cast: CastMember[];
  creator: string;
  thumbnailUrl: string;
  bannerUrl: string;
  trailerUrl: string | null;
  averageRating: number;
  ratingCount: number;
  status: string;
  tags: string[];
  seasons: Season[];
  totalSeasons: number;
  totalEpisodes: number;
  createdAt: string;
  updatedAt: string;
}

export interface Genre {
  id: string;
  name: string;
  slug: string;
  description: string;
  iconUrl: string | null;
  contentCount: number;
}

export interface ContentSummary {
  id: string;
  title: string;
  description: string;
  releaseYear: number;
  contentType: ContentType;
  maturityRating: string;
  genres: string[];
  thumbnailUrl: string;
  bannerUrl: string;
  averageRating: number | null;
  status: string;
}

export interface CatalogParams {
  page?: number;
  pageSize?: number;
  genre?: string;
  sort?: string;
  year?: number;
}
