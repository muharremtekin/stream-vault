'use client';

import { useCallback, useEffect, useRef } from 'react';

import { toast } from 'sonner';

import { useNotifications } from '@/lib/hooks/use-notifications';
import { useAuthStore } from '@/lib/stores/auth-store';
import { useNotificationStore } from '@/lib/stores/notification-store';
import { getNotificationIcon } from '@/lib/utils/notification-icons';
import type { Notification } from '@/lib/types/notification';

interface NotificationProviderProps {
  children: React.ReactNode;
}

export function NotificationProvider({ children }: NotificationProviderProps) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const prevTokenRef = useRef(accessToken);

  const handleNewNotification = useCallback(
    (notification: Notification) => {
      const Icon = getNotificationIcon(notification.type);
      toast(notification.title, {
        description: notification.body,
        icon: <Icon className="h-4 w-4" />,
      });
    },
    []
  );

  useEffect(() => {
    if (prevTokenRef.current && !accessToken) {
      useNotificationStore.getState().reset();
    }
    prevTokenRef.current = accessToken;
  }, [accessToken]);

  if (accessToken) {
    return (
      <NotificationConnection onNewNotification={handleNewNotification}>
        {children}
      </NotificationConnection>
    );
  }

  return <>{children}</>;
}

function NotificationConnection({
  children,
  onNewNotification,
}: {
  children: React.ReactNode;
  onNewNotification: (n: Notification) => void;
}) {
  useNotifications({ onNewNotification });
  return <>{children}</>;
}
