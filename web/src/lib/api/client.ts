import axios from 'axios';
import type { AxiosError, InternalAxiosRequestConfig } from 'axios';

import { API_URL } from '@/lib/utils/constants';
import type { ApiError } from '@/lib/types/common';

export const apiClient = axios.create({
  baseURL: API_URL,
  timeout: 10_000,
  headers: { 'Content-Type': 'application/json' },
});

// ---------------------------------------------------------------------------
// Request interceptor: attach Bearer token + profile ID from auth store
// ---------------------------------------------------------------------------
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  // Dynamic require to avoid circular dependency (client ↔ auth-store)
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { useAuthStore } = require('@/lib/stores/auth-store') as {
    useAuthStore: {
      getState: () => {
        accessToken: string | null;
        activeProfile: { id: string } | null;
      };
    };
  };

  const { accessToken, activeProfile } = useAuthStore.getState();

  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  if (activeProfile) {
    config.headers['X-Profile-Id'] = activeProfile.id;
  }

  return config;
});

// ---------------------------------------------------------------------------
// Response interceptor: 401 → refresh token → retry original request
// ---------------------------------------------------------------------------

/** Queue of requests waiting for the refresh to complete */
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null): void {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else if (token) {
      prom.resolve(token);
    }
  });
  failedQueue = [];
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined;

    // Only handle 401 on non-retry requests that have config
    if (
      error.response?.status !== 401 ||
      !originalRequest ||
      originalRequest._retry
    ) {
      return Promise.reject(error);
    }

    // If a refresh is already in progress, queue this request
    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then((token) => {
        originalRequest.headers.Authorization = `Bearer ${token}`;
        return apiClient(originalRequest);
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      // eslint-disable-next-line @typescript-eslint/no-require-imports
      const { useAuthStore } = require('@/lib/stores/auth-store') as {
        useAuthStore: {
          getState: () => {
            refreshToken: string | null;
            setTokens: (at: string, rt: string) => void;
            logout: () => void;
          };
        };
      };

      const { refreshToken } = useAuthStore.getState();

      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      // Use a plain axios call (not apiClient) to avoid interceptor recursion
      const { data } = await axios.post<{
        accessToken: string;
        refreshToken: string;
      }>(`${API_URL}/api/auth/refresh`, { refreshToken });

      useAuthStore.getState().setTokens(data.accessToken, data.refreshToken);

      processQueue(null, data.accessToken);

      originalRequest.headers.Authorization = `Bearer ${data.accessToken}`;
      return apiClient(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError, null);

      // eslint-disable-next-line @typescript-eslint/no-require-imports
      const { useAuthStore } = require('@/lib/stores/auth-store') as {
        useAuthStore: {
          getState: () => { logout: () => void };
        };
      };

      useAuthStore.getState().logout();

      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }

      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);
