'use client';

import { useTranslations } from 'next-intl';

interface ErrorPageProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function ErrorPage({ reset }: ErrorPageProps) {
  const t = useTranslations('error');

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 px-4">
      <div className="text-center">
        <h1 className="mb-2 text-2xl font-bold text-foreground">
          {t('title')}
        </h1>
        <p className="text-muted-foreground">{t('description')}</p>
      </div>
      <div className="flex gap-4">
        <button
          onClick={reset}
          className="rounded-md bg-primary px-6 py-2 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t('tryAgain')}
        </button>
        <a
          href="/"
          className="rounded-md border border-border bg-card px-6 py-2 text-sm font-semibold text-foreground transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t('goHome')}
        </a>
      </div>
    </div>
  );
}
