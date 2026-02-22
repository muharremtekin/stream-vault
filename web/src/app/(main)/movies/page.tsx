import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

import { MoviesContent } from '@/components/catalog/movies-content';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('pages.movies');
  return {
    title: t('title'),
    description: t('metaDescription'),
  };
}

export default async function MoviesPage() {
  const t = await getTranslations('pages.movies');

  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      <h1 className="mb-6 text-2xl font-bold text-foreground sm:text-3xl">
        {t('title')}
      </h1>
      <MoviesContent />
    </div>
  );
}
