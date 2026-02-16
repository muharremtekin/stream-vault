import axios from 'axios';

import type { ApiError } from '@/lib/types/common';

/**
 * Extracts a human-readable error message from various error shapes
 * returned by backend services (Axios errors, ApiError, generic errors).
 */
export function extractErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as ApiError | undefined;

    if (data?.message) return data.message;
    if (data?.error) return data.error;
    if (data?.errors && data.errors.length > 0) {
      return data.errors[0].message;
    }

    if (error.code === 'ECONNABORTED') {
      return 'Request timed out. Please try again.';
    }
    if (!error.response) {
      return 'Network error. Please check your connection.';
    }

    return error.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return 'An unexpected error occurred.';
}
