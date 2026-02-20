'use client';

import { useState, useRef, useEffect, useCallback } from 'react';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Bell, CheckCheck } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import { useNotificationStore } from '@/lib/stores/notification-store';
import {
  markAsRead as markAsReadApi,
  markAllAsRead as markAllAsReadApi,
} from '@/lib/api/notification';
import { showSuccessToast, showErrorToast } from '@/components/ui/toast';
import { extractErrorMessage } from '@/lib/utils/error';
import { NotificationItem } from '@/components/notification/notification-item';
import type { Notification } from '@/lib/types/notification';

const MAX_DROPDOWN_ITEMS = 5;

export function NotificationBell() {
  const t = useTranslations('layout.notificationBell');
  const tApi = useTranslations('api');
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const notifications = useNotificationStore((s) => s.notifications);
  const unreadCount = useNotificationStore((s) => s.unreadCount);
  const markAsReadStore = useNotificationStore((s) => s.markAsRead);
  const markAllAsReadStore = useNotificationStore((s) => s.markAllAsRead);
  const setUnreadCount = useNotificationStore((s) => s.setUnreadCount);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const handleItemClick = useCallback(
    async (notification: Notification) => {
      try {
        if (!notification.read) {
          await markAsReadApi(notification.id);
          markAsReadStore(notification.id);
          setUnreadCount(
            Math.max(0, useNotificationStore.getState().unreadCount - 1)
          );
        }
        if (notification.action) {
          router.push(notification.action);
        }
      } catch (error) {
        showErrorToast(extractErrorMessage(error));
      }
      setIsOpen(false);
    },
    [markAsReadStore, setUnreadCount, router]
  );

  const handleMarkAllAsRead = useCallback(async () => {
    try {
      await markAllAsReadApi();
      markAllAsReadStore();
      showSuccessToast(tApi('allNotificationsRead'));
    } catch (error) {
      showErrorToast(extractErrorMessage(error));
    }
  }, [markAllAsReadStore, tApi]);

  const displayedNotifications = notifications.slice(0, MAX_DROPDOWN_ITEMS);

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen((prev) => !prev)}
        aria-label={t('notifications')}
        className="relative rounded-full p-2 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        <Bell className="h-5 w-5" />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 rounded-md border border-border bg-card shadow-lg">
          <div className="flex items-center justify-between border-b border-border px-4 py-3">
            <span className="text-sm font-semibold text-foreground">
              {t('notifications')}
            </span>
            {unreadCount > 0 && (
              <button
                onClick={handleMarkAllAsRead}
                aria-label={t('markAllRead')}
                className="rounded p-1 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                <CheckCheck className="h-4 w-4" />
              </button>
            )}
          </div>

          {displayedNotifications.length > 0 ? (
            <div className="max-h-80 divide-y divide-border overflow-y-auto">
              {displayedNotifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onClick={handleItemClick}
                  isCompact
                />
              ))}
            </div>
          ) : (
            <div className="px-4 py-6 text-center text-sm text-muted-foreground">
              {t('noNotifications')}
            </div>
          )}

          <div className="border-t border-border px-4 py-2">
            <Link
              href="/account/notifications"
              onClick={() => setIsOpen(false)}
              className="block text-center text-sm text-primary transition-colors hover:text-primary/80"
            >
              {t('viewAll')}
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
