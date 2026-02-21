import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { AdminDashboard } from '@/components/admin/admin-dashboard';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('admin.dashboard');
  return {
    title: t('pageTitle'),
  };
}

export default async function AdminDashboardPage() {
  const t = await getTranslations('admin.dashboard');

  return (
    <>
      <h1 className="mb-8 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <AdminDashboard />
    </>
  );
}
