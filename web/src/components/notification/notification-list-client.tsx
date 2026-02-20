'use client';

import { useState, useCallback } from 'react';

import { useRouter } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Bell, CheckCheck, ChevronLeft, ChevronRight } from 'lucide-react';
import { useTranslations } from 'next-intl';

import {
  getNotifications,
  markAsRead as markAsReadApi,
  markAllAsRead as markAllAsReadApi,
} from '@/lib/api/notification';
import { useNotificationStore } from '@/lib/stores/notification-store';
import { showSuccessToast, showErrorToast } from '@/components/ui/toast';
import { extractErrorMessage } from '@/lib/utils/error';
import { NotificationItem } from '@/components/notification/notification-item';
import { cn } from '@/lib/utils/cn';
import type { Notification } from '@/lib/types/notification';

const PAGE_SIZE = 20;

export function NotificationListClient() {
  const t = useTranslations('notifications');
  const tApi = useTranslations('api');
  const router = useRouter();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [isUnreadOnly, setIsUnreadOnly] = useState(false);

  const markAsReadStore = useNotificationStore((s) => s.markAsRead);
  const markAllAsReadStore = useNotificationStore((s) => s.markAllAsRead);
  const setUnreadCount = useNotificationStore((s) => s.setUnreadCount);

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', 'list', { page, pageSize: PAGE_SIZE, unreadOnly: isUnreadOnly }],
    queryFn: () => getNotifications({ page, pageSize: PAGE_SIZE, unreadOnly: isUnreadOnly }),
    staleTime: 30_000,
  });

  const handleItemClick = useCallback(
    async (notification: Notification) => {
      try {
        if (!notification.read) {
          await markAsReadApi(notification.id);
          markAsReadStore(notification.id);
          setUnreadCount(
            Math.max(0, useNotificationStore.getState().unreadCount - 1)
          );
          queryClient.invalidateQueries({ queryKey: ['notifications'] });
        }
        if (notification.action) {
          router.push(notification.action);
        }
      } catch (error) {
        showErrorToast(extractErrorMessage(error));
      }
    },
    [markAsReadStore, setUnreadCount, queryClient, router]
  );

  const handleMarkAllAsRead = useCallback(async () => {
    try {
      await markAllAsReadApi();
      markAllAsReadStore();
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      showSuccessToast(tApi('allNotificationsRead'));
    } catch (error) {
      showErrorToast(extractErrorMessage(error));
    }
  }, [markAllAsReadStore, queryClient, tApi]);

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={`skeleton-${i}`} className="flex items-start gap-3 rounded-lg px-4 py-4">
            <div className="h-10 w-10 shrink-0 animate-pulse rounded-full bg-muted" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
              <div className="h-3 w-1/2 animate-pulse rounded bg-muted" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  const items = data?.items ?? [];
  const totalPages = data?.totalPages ?? 1;

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <label className="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            checked={isUnreadOnly}
            onChange={(e) => {
              setIsUnreadOnly(e.target.checked);
              setPage(1);
            }}
            className="h-4 w-4 rounded border-border bg-card accent-primary"
          />
          {t('unreadOnly')}
        </label>
        {(data?.unreadCount ?? 0) > 0 && (
          <button
            onClick={handleMarkAllAsRead}
            className="flex items-center gap-1.5 rounded px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            <CheckCheck className="h-4 w-4" />
            {t('markAsRead')}
          </button>
        )}
      </div>

      {items.length > 0 ? (
        <div className="divide-y divide-border rounded-lg border border-border">
          {items.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onClick={handleItemClick}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border border-border py-16 text-center">
          <Bell className="mb-4 h-12 w-12 text-muted-foreground/50" />
          <p className="text-lg font-medium text-foreground">{t('empty')}</p>
          <p className="mt-1 text-sm text-muted-foreground">{t('emptyDescription')}</p>
        </div>
      )}

      {totalPages > 1 && (
        <div className="mt-6 flex items-center justify-center gap-4">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            aria-label={t('previousPage')}
            className={cn(
              'rounded p-2 transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
              page <= 1
                ? 'cursor-not-allowed text-muted-foreground/30'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <ChevronLeft className="h-5 w-5" />
          </button>
          <span className="text-sm text-muted-foreground">
            {t('page', { current: page, total: totalPages })}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            aria-label={t('nextPage')}
            className={cn(
              'rounded p-2 transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
              page >= totalPages
                ? 'cursor-not-allowed text-muted-foreground/30'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <ChevronRight className="h-5 w-5" />
          </button>
        </div>
      )}
    </div>
  );
}
