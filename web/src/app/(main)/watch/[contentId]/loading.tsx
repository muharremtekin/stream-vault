import { getTranslations } from 'next-intl/server';

export default async function WatchLoading() {
  const t = await getTranslations('player');

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black">
      <span className="sr-only">{t('loading')}</span>
      <div className="h-12 w-12 animate-spin rounded-full border-4 border-white border-t-transparent" />
    </div>
  );
}
