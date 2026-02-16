export interface RecommendedItem {
  contentId: string;
  title: string;
  thumbnailUrl?: string;
  contentType: string;
  releaseYear: number;
  averageRating: number;
  score: number;
  algorithm: string;
  reason: string;
}

export interface SimilarItem {
  contentId: string;
  title: string;
  thumbnailUrl?: string;
  contentType: string;
  releaseYear: number;
  averageRating: number;
  similarityScore: number;
}

export interface HomePageSection {
  sectionType: string;
  title: string;
  items: RecommendedItem[];
}

export interface RecommendationsResponse {
  items: RecommendedItem[];
}

export interface SimilarResponse {
  items: SimilarItem[];
}

export interface HomePageResponse {
  sections: HomePageSection[];
}

export interface FeedbackRequest {
  contentId: string;
  feedbackType: string;
}
