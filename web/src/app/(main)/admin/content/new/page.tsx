import { getTranslations } from 'next-intl/server';

import { ContentForm } from '@/components/admin/content-form';

export async function generateMetadata() {
  const t = await getTranslations('admin.contentForm');
  return { title: t('pageTitle') };
}

export default async function AdminContentNewPage() {
  const t = await getTranslations('admin.contentForm');

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <div className="rounded-lg border border-border bg-card p-6">
        <ContentForm />
      </div>
    </div>
  );
}
