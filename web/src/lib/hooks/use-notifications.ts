'use client';

import { useEffect, useRef } from 'react';

import { useQuery } from '@tanstack/react-query';

import {
  getNotifications,
  markAsRead as markAsReadApi,
  markAllAsRead as markAllAsReadApi,
} from '@/lib/api/notification';
import { useNotificationStore } from '@/lib/stores/notification-store';
import { useAuthStore } from '@/lib/stores/auth-store';
import { WS_URL } from '@/lib/utils/constants';
import type { Notification } from '@/lib/types/notification';

const WS_RECONNECT_DELAY = 5_000;

export function useNotifications() {
  const accessToken = useAuthStore((s) => s.accessToken);
  const setNotifications = useNotificationStore((s) => s.setNotifications);
  const addNotification = useNotificationStore((s) => s.addNotification);
  const setUnreadCount = useNotificationStore((s) => s.setUnreadCount);
  const setWsConnected = useNotificationStore((s) => s.setWsConnected);
  const markAsReadStore = useNotificationStore((s) => s.markAsRead);
  const markAllAsReadStore = useNotificationStore((s) => s.markAllAsRead);
  const notifications = useNotificationStore((s) => s.notifications);
  const unreadCount = useNotificationStore((s) => s.unreadCount);
  const wsConnected = useNotificationStore((s) => s.wsConnected);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const notificationsQuery = useQuery({
    queryKey: ['notifications', 'list'],
    queryFn: () => getNotifications({ page: 1, pageSize: 20 }),
    staleTime: 60_000,
    enabled: !!accessToken,
  });

  // Sync fetched data to store
  useEffect(() => {
    if (notificationsQuery.data) {
      setNotifications(notificationsQuery.data.items);
      setUnreadCount(notificationsQuery.data.unreadCount);
    }
  }, [notificationsQuery.data, setNotifications, setUnreadCount]);

  // WebSocket connection managed entirely inside useEffect
  useEffect(() => {
    if (!accessToken) return;

    let isCancelled = false;

    function connect() {
      if (isCancelled) return;
      if (wsRef.current?.readyState === WebSocket.OPEN) return;

      const ws = new WebSocket(
        `${WS_URL}/ws/notifications?token=${accessToken}`
      );
      wsRef.current = ws;

      ws.onopen = () => {
        if (!isCancelled) setWsConnected(true);
      };

      ws.onmessage = (event: MessageEvent) => {
        if (isCancelled) return;
        try {
          const notification = JSON.parse(
            event.data as string
          ) as Notification;
          addNotification(notification);
          setUnreadCount(
            useNotificationStore.getState().unreadCount + 1
          );
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onclose = () => {
        if (!isCancelled) {
          setWsConnected(false);
          reconnectRef.current = setTimeout(connect, WS_RECONNECT_DELAY);
        }
      };

      ws.onerror = () => ws.close();
    }

    connect();

    return () => {
      isCancelled = true;
      wsRef.current?.close();
      if (reconnectRef.current) clearTimeout(reconnectRef.current);
    };
  }, [accessToken, setWsConnected, addNotification, setUnreadCount]);

  const markAsRead = async (id: string) => {
    await markAsReadApi(id);
    markAsReadStore(id);
    setUnreadCount(
      Math.max(0, useNotificationStore.getState().unreadCount - 1)
    );
  };

  const markAllAsRead = async () => {
    await markAllAsReadApi();
    markAllAsReadStore();
  };

  return {
    notifications,
    unreadCount,
    wsConnected,
    isLoading: notificationsQuery.isLoading,
    markAsRead,
    markAllAsRead,
    refetch: notificationsQuery.refetch,
  };
}
