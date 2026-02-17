const TOKEN_COOKIE = 'sv-access-token';
const REFRESH_COOKIE = 'sv-refresh-token';

export function setAuthCookies(
  accessToken: string,
  refreshToken: string
): void {
  document.cookie = `${TOKEN_COOKIE}=${accessToken}; path=/; max-age=86400; SameSite=Lax`;
  document.cookie = `${REFRESH_COOKIE}=${refreshToken}; path=/; max-age=604800; SameSite=Lax`;
}

export function clearAuthCookies(): void {
  document.cookie = `${TOKEN_COOKIE}=; path=/; max-age=0`;
  document.cookie = `${REFRESH_COOKIE}=; path=/; max-age=0`;
}

export function getTokenCookieName(): string {
  return TOKEN_COOKIE;
}
