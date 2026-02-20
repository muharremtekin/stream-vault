'use client';

import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import { formatRelativeTime } from '@/lib/utils/format';
import { getNotificationIcon } from '@/lib/utils/notification-icons';
import type { Notification } from '@/lib/types/notification';

interface NotificationItemProps {
  notification: Notification;
  onClick: (notification: Notification) => void;
  isCompact?: boolean;
}

export function NotificationItem({
  notification,
  onClick,
  isCompact = false,
}: NotificationItemProps) {
  const t = useTranslations('notifications');
  const Icon = getNotificationIcon(notification.type);
  const maxBodyLength = isCompact ? 50 : 100;
  const truncatedBody =
    notification.body.length > maxBodyLength
      ? `${notification.body.slice(0, maxBodyLength)}...`
      : notification.body;

  return (
    <button
      onClick={() => onClick(notification)}
      className={cn(
        'flex w-full items-start gap-3 text-left transition-colors',
        isCompact ? 'px-4 py-3' : 'rounded-lg px-4 py-4',
        notification.read
          ? 'hover:bg-muted/50'
          : 'bg-muted/30 hover:bg-muted/50'
      )}
      aria-label={t('markAsRead')}
    >
      <div
        className={cn(
          'flex shrink-0 items-center justify-center rounded-full',
          isCompact ? 'h-8 w-8' : 'h-10 w-10',
          'bg-muted'
        )}
      >
        <Icon className={cn(isCompact ? 'h-4 w-4' : 'h-5 w-5', 'text-muted-foreground')} />
      </div>

      <div className="min-w-0 flex-1">
        <p
          className={cn(
            'truncate font-medium text-foreground',
            isCompact ? 'text-xs' : 'text-sm'
          )}
        >
          {notification.title}
        </p>
        <p
          className={cn(
            'text-muted-foreground',
            isCompact ? 'text-xs' : 'text-sm'
          )}
        >
          {truncatedBody}
        </p>
        <p className="mt-1 text-xs text-muted-foreground">
          {formatRelativeTime(notification.createdAt)}
        </p>
      </div>

      {!notification.read && (
        <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-primary" />
      )}
    </button>
  );
}
