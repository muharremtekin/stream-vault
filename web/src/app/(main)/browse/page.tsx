import { getTranslations } from 'next-intl/server';

export default async function BrowsePage() {
  const t = await getTranslations('browse');

  return (
    <div className="flex min-h-[calc(100vh-4rem)] items-center justify-center">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-foreground">{t('home')}</h1>
        <p className="mt-2 text-muted-foreground">{t('trending')}</p>
      </div>
    </div>
  );
}
