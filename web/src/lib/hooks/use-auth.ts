'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { login as loginApi, register as registerApi } from '@/lib/api/auth';
import { useAuthStore } from '@/lib/stores/auth-store';
import { setAuthCookies, clearAuthCookies } from '@/lib/utils/cookies';
import { extractErrorMessage } from '@/lib/utils/error';
import type { LoginRequest, RegisterRequest, AuthResponse } from '@/lib/types/auth';

export function useAuth() {
  const queryClient = useQueryClient();
  const setAuth = useAuthStore((s) => s.setAuth);
  const logoutStore = useAuthStore((s) => s.logout);
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.accessToken !== null);

  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) => loginApi(data),
    onSuccess: (response: AuthResponse) => {
      setAuth(response.user, response.accessToken, response.refreshToken);
      setAuthCookies(response.accessToken, response.refreshToken);
    },
  });

  const registerMutation = useMutation({
    mutationFn: (data: RegisterRequest) => registerApi(data),
    onSuccess: (response: AuthResponse) => {
      setAuth(response.user, response.accessToken, response.refreshToken);
      setAuthCookies(response.accessToken, response.refreshToken);
    },
  });

  const logout = () => {
    clearAuthCookies();
    logoutStore();
    queryClient.clear();
  };

  return {
    user,
    isAuthenticated,
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    logout,
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    loginError: loginMutation.error
      ? extractErrorMessage(loginMutation.error)
      : null,
    registerError: registerMutation.error
      ? extractErrorMessage(registerMutation.error)
      : null,
  };
}
