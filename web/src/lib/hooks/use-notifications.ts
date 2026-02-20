'use client';

import { useEffect, useRef, useCallback } from 'react';

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

const MAX_RECONNECT_DELAY = 30_000;
const BASE_RECONNECT_DELAY = 1_000;

interface UseNotificationsOptions {
  onNewNotification?: (notification: Notification) => void;
}

export function useNotifications(options?: UseNotificationsOptions) {
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
  const reconnectAttemptRef = useRef(0);
  const onNewNotificationRef = useRef(options?.onNewNotification);

  useEffect(() => {
    onNewNotificationRef.current = options?.onNewNotification;
  }, [options?.onNewNotification]);

  const notificationsQuery = useQuery({
    queryKey: ['notifications', 'list'],
    queryFn: () => getNotifications({ page: 1, pageSize: 20 }),
    staleTime: 60_000,
    enabled: !!accessToken,
  });

  useEffect(() => {
    if (notificationsQuery.data) {
      setNotifications(notificationsQuery.data.items);
      setUnreadCount(notificationsQuery.data.unreadCount);
    }
  }, [notificationsQuery.data, setNotifications, setUnreadCount]);

  useEffect(() => {
    if (!accessToken) return;

    let isCancelled = false;

    function getReconnectDelay(): number {
      const delay = Math.min(
        BASE_RECONNECT_DELAY * Math.pow(2, reconnectAttemptRef.current),
        MAX_RECONNECT_DELAY
      );
      return delay;
    }

    function connect() {
      if (isCancelled) return;
      if (wsRef.current?.readyState === WebSocket.OPEN) return;

      const ws = new WebSocket(
        `${WS_URL}/ws/notifications?token=${accessToken}`
      );
      wsRef.current = ws;

      ws.onopen = () => {
        if (!isCancelled) {
          setWsConnected(true);
          reconnectAttemptRef.current = 0;
        }
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
          onNewNotificationRef.current?.(notification);
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onclose = () => {
        if (!isCancelled) {
          setWsConnected(false);
          const delay = getReconnectDelay();
          reconnectAttemptRef.current += 1;
          reconnectRef.current = setTimeout(connect, delay);
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

  const markAsRead = useCallback(async (id: string) => {
    await markAsReadApi(id);
    markAsReadStore(id);
    setUnreadCount(
      Math.max(0, useNotificationStore.getState().unreadCount - 1)
    );
  }, [markAsReadStore, setUnreadCount]);

  const markAllAsRead = useCallback(async () => {
    await markAllAsReadApi();
    markAllAsReadStore();
  }, [markAllAsReadStore]);

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
