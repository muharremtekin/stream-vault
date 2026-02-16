export interface Highlight {
  title?: string;
  description?: string;
}

export interface SearchHit {
  contentId: string;
  title: string;
  description: string;
  contentType: string;
  thumbnailUrl: string;
  releaseYear: number;
  averageRating: number;
  maturityRating: string;
  genres: string[];
  highlights?: Highlight;
}

export interface FacetBucket {
  key: string;
  docCount: number;
}

export interface Facets {
  genreFacets: FacetBucket[];
  yearFacets: FacetBucket[];
  maturityFacets: FacetBucket[];
  contentTypeFacets: FacetBucket[];
}

export interface SearchResponse {
  items: SearchHit[];
  facets?: Facets;
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
}

export interface SearchParams {
  q?: string;
  genres?: string;
  content_type?: string;
  year_from?: number;
  year_to?: number;
  min_rating?: number;
  maturity_ratings?: string;
  sort?: string;
  order?: string;
  page?: number;
  pageSize?: number;
}

export interface AutocompleteSuggestion {
  contentId: string;
  title: string;
  contentType: string;
  thumbnailUrl: string;
  releaseYear: number;
}

export interface AutocompleteResponse {
  suggestions: AutocompleteSuggestion[];
}

export interface TrendingItem {
  contentId: string;
  title: string;
  thumbnailUrl: string;
  contentType: string;
  releaseYear: number;
  averageRating: number;
  rank: number;
  rankChange: number;
  viewCount: number;
}

export interface TrendingResponse {
  items: TrendingItem[];
}
