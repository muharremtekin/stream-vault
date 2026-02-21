import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { AccountOverview } from '@/components/account/account-overview';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('account');
  return {
    title: t('pageTitle'),
  };
}

export default async function AccountPage() {
  const t = await getTranslations('account');

  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <h1 className="mb-8 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <AccountOverview />
    </div>
  );
}
