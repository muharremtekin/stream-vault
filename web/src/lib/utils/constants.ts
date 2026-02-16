export const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
export const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8081';

/** Max video quality (height) per subscription tier */
export const TIER_MAX_QUALITY: Record<string, number> = {
  Free: 480,
  Basic: 720,
  Standard: 1080,
  Premium: 4320,
  Admin: 4320,
};

/** Max concurrent streams per subscription tier */
export const TIER_MAX_STREAMS: Record<string, number> = {
  Free: 1,
  Basic: 1,
  Standard: 2,
  Premium: 4,
  Admin: 4,
};

/** Max profiles per account */
export const MAX_PROFILES = 5;

/** Progress save interval in ms */
export const PROGRESS_SAVE_INTERVAL = 15_000;

/** Autocomplete debounce in ms */
export const SEARCH_DEBOUNCE = 300;

/** Min characters for autocomplete */
export const SEARCH_MIN_CHARS = 2;
