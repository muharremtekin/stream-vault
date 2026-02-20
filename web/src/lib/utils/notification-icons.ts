import { Bell, CreditCard, Film, Tv } from 'lucide-react';

import type { LucideIcon } from 'lucide-react';

const NOTIFICATION_TYPE_ICONS: Record<string, LucideIcon> = {
  subscription: CreditCard,
  encoding: Film,
  content: Tv,
};

export function getNotificationIcon(type: string): LucideIcon {
  return NOTIFICATION_TYPE_ICONS[type] ?? Bell;
}
