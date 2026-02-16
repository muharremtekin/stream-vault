'use client';

import { create } from 'zustand';

import type { Notification } from '@/lib/types/notification';

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  wsConnected: boolean;

  setNotifications: (notifications: Notification[]) => void;
  addNotification: (notification: Notification) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  setUnreadCount: (count: number) => void;
  setWsConnected: (connected: boolean) => void;
  reset: () => void;
}

const initialState = {
  notifications: [] as Notification[],
  unreadCount: 0,
  wsConnected: false,
};

export const useNotificationStore = create<NotificationState>()((set) => ({
  ...initialState,

  setNotifications: (notifications) => set({ notifications }),

  addNotification: (notification) =>
    set((state) => ({
      notifications: [notification, ...state.notifications],
    })),

  markAsRead: (id) =>
    set((state) => ({
      notifications: state.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n
      ),
    })),

  markAllAsRead: () =>
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, read: true })),
      unreadCount: 0,
    })),

  setUnreadCount: (unreadCount) => set({ unreadCount }),

  setWsConnected: (wsConnected) => set({ wsConnected }),

  reset: () => set(initialState),
}));
