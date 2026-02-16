import { apiClient } from '@/lib/api/client';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  RefreshTokenRequest,
  Profile,
  CreateProfileRequest,
} from '@/lib/types/auth';

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>(
    '/api/auth/login',
    data
  );
  return response.data;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>(
    '/api/auth/register',
    data
  );
  return response.data;
}

export async function refreshToken(
  data: RefreshTokenRequest
): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>(
    '/api/auth/refresh',
    data
  );
  return response.data;
}

export async function getProfiles(userId: string): Promise<Profile[]> {
  const response = await apiClient.get<Profile[]>(
    `/api/profile/user/${userId}`
  );
  return response.data;
}

export async function createProfile(
  data: CreateProfileRequest
): Promise<Profile> {
  const response = await apiClient.post<Profile>('/api/profile', data);
  return response.data;
}
