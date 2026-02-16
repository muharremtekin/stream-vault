import { getTranslations } from 'next-intl/server';

export default async function LoginPage() {
  const t = await getTranslations('auth');

  return (
    <div className="w-full max-w-md rounded-lg bg-card p-8">
      <h1 className="mb-6 text-2xl font-bold text-foreground">
        {t('loginTitle')}
      </h1>
      <p className="text-muted-foreground">
        {t('loginTitle')}
      </p>
    </div>
  );
}
