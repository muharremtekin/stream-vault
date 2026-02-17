import { getTranslations } from 'next-intl/server';

import WatchlistGrid from '@/components/browse/watchlist-grid';

export default async function MyListPage() {
  const t = await getTranslations('browse');

  return (
    <main className="min-h-screen bg-background px-4 sm:px-8 lg:px-12 pt-24 pb-16">
      <h1 className="mb-6 text-2xl font-bold text-foreground sm:text-3xl">
        {t('myList')}
      </h1>
      <WatchlistGrid />
    </main>
  );
}
