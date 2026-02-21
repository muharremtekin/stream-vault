import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { ContentManagement } from '@/components/admin/content-management';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('admin.content');
  return {
    title: t('pageTitle'),
  };
}

export default async function ContentManagementPage() {
  const t = await getTranslations('admin.content');

  return (
    <>
      <h1 className="mb-8 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <ContentManagement />
    </>
  );
}
