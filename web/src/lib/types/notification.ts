export interface ChannelStatus {
  channel: string;
  status: string;
  sentAt?: string;
}

export interface Notification {
  id: string;
  userId: string;
  type: string;
  category: string;
  title: string;
  body: string;
  icon?: string;
  action?: string;
  read: boolean;
  channels: ChannelStatus[];
  createdAt: string;
  expiresAt: string;
}

export interface NotificationListResponse {
  items: Notification[];
  unreadCount: number;
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ChannelPreference {
  enabled: boolean;
  categories?: string[];
}

export interface NotificationPreferences {
  id: string;
  userId: string;
  email: ChannelPreference;
  push: ChannelPreference;
  inApp: ChannelPreference;
  updatedAt: string;
}

export interface UpdatePreferencesRequest {
  email?: ChannelPreference;
  push?: ChannelPreference;
  inApp?: ChannelPreference;
}
