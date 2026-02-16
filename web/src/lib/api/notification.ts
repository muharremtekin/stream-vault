import { apiClient } from '@/lib/api/client';
import type {
  NotificationListResponse,
  NotificationPreferences,
  UpdatePreferencesRequest,
} from '@/lib/types/notification';

interface NotificationListParams {
  page?: number;
  pageSize?: number;
  unreadOnly?: boolean;
}

export async function getNotifications(
  params?: NotificationListParams
): Promise<NotificationListResponse> {
  const response = await apiClient.get<NotificationListResponse>(
    '/api/notifications',
    { params }
  );
  return response.data;
}

export async function markAsRead(id: string): Promise<void> {
  await apiClient.post(`/api/notifications/${id}/read`);
}

export async function markAllAsRead(): Promise<void> {
  await apiClient.post('/api/notifications/read-all');
}

export async function getPreferences(): Promise<NotificationPreferences> {
  const response = await apiClient.get<NotificationPreferences>(
    '/api/notifications/preferences'
  );
  return response.data;
}

export async function updatePreferences(
  data: UpdatePreferencesRequest
): Promise<NotificationPreferences> {
  const response = await apiClient.put<NotificationPreferences>(
    '/api/notifications/preferences',
    data
  );
  return response.data;
}
