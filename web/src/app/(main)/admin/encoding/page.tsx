import { getTranslations } from 'next-intl/server';

import { EncodingStatus } from '@/components/admin/encoding-status';

export async function generateMetadata() {
  const t = await getTranslations('admin.encoding');
  return { title: t('pageTitle') };
}

export default async function AdminEncodingPage() {
  const t = await getTranslations('admin.encoding');

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-foreground">{t('pageTitle')}</h1>
      <EncodingStatus />
    </div>
  );
}
