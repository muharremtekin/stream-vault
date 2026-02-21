import { apiClient } from '@/lib/api/client';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  RefreshTokenRequest,
  Profile,
  CreateProfileRequest,
  UpdateProfileRequest,
} from '@/lib/types/auth';

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>(
    '/api/auth/login',
    data
  );
  return response.data;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const { confirmPassword: _unused, ...payload } = data;
  const response = await apiClient.post<AuthResponse>(
    '/api/auth/register',
    payload
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

export async function updateProfile(
  id: string,
  data: UpdateProfileRequest
): Promise<Profile> {
  const response = await apiClient.put<Profile>(`/api/profile/${id}`, data);
  return response.data;
}

export async function deleteProfile(id: string): Promise<void> {
  await apiClient.delete(`/api/profile/${id}`);
}
