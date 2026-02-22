/** Generic paginated response matching backend PagedResult<T> */
export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
}

/** Standard API error structure (varies by backend service) */
export interface ApiError {
  error?: string;
  message?: string;
  code?: string;
  errors?: FieldError[];
  status?: number;
}

export interface FieldError {
  field: string;
  message: string;
}

/** Generic single-item API response wrapper */
export interface ApiResponse<T> {
  data: T;
}

/** Content type (Movie or Series) */
export const CONTENT_TYPES = {
  Movie: 'Movie',
  Series: 'Series',
} as const;

export type ContentType = (typeof CONTENT_TYPES)[keyof typeof CONTENT_TYPES];

/** Video status from catalog service (matches CatalogService.Domain.Enums.VideoStatus) */
export const VIDEO_STATUSES = {
  NotUploaded: 'NotUploaded',
  Uploading: 'Uploading',
  Queued: 'Queued',
  Encoding: 'Encoding',
  Ready: 'Ready',
  Error: 'Error',
} as const;

export type VideoStatus = (typeof VIDEO_STATUSES)[keyof typeof VIDEO_STATUSES];

/** Subscription tier (matches UserService.Domain.Enums.SubscriptionTier) */
export const SUBSCRIPTION_TIERS = {
  Free: 'Free',
  Basic: 'Basic',
  Standard: 'Standard',
  Premium: 'Premium',
  Admin: 'Admin',
} as const;

export type SubscriptionTier =
  (typeof SUBSCRIPTION_TIERS)[keyof typeof SUBSCRIPTION_TIERS];
