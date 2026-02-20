import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { NotificationListClient } from '@/components/notification/notification-list-client';
import { NotificationPrefs } from '@/components/notification/notification-prefs';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('notifications');
  return {
    title: t('pageTitle'),
  };
}

export default async function NotificationsPage() {
  const t = await getTranslations('notifications');

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <h1 className="mb-8 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <NotificationListClient />
      <hr className="my-8 border-border" />
      <NotificationPrefs />
    </div>
  );
}
