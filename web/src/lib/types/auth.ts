export interface User {
  id: string;
  email: string;
  role: string;
  profileCount: number;
}

export interface Profile {
  id: string;
  name: string;
  icon: string;
  isKids: boolean;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  confirmPassword: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface CreateProfileRequest {
  userId: string;
  name: string;
  icon: string;
  isKids: boolean;
}
