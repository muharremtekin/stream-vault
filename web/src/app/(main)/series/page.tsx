import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { SeriesGrid } from '@/components/catalog/series-grid';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('pages.series');
  return { title: t('title') };
}

export default async function SeriesPage() {
  const t = await getTranslations('pages.series');

  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      <h1 className="mb-1 text-3xl font-bold text-foreground">
        {t('title')}
      </h1>
      <SeriesGrid />
    </div>
  );
}
