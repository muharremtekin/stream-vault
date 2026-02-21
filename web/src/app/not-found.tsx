import Link from 'next/link';

import { getTranslations } from 'next-intl/server';

export default async function NotFound() {
  const t = await getTranslations('error');

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-background px-4">
      <div className="text-center">
        <p className="text-8xl font-bold text-primary">404</p>
        <h1 className="mt-4 text-2xl font-bold text-foreground">
          {t('notFoundTitle')}
        </h1>
        <p className="mt-2 text-muted-foreground">
          {t('notFoundDescription')}
        </p>
      </div>
      <Link
        href="/browse"
        className="rounded-md bg-primary px-6 py-2 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        {t('backToHome')}
      </Link>
    </div>
  );
}
