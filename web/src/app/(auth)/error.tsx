'use client';

import { useTranslations } from 'next-intl';

interface AuthErrorPageProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function AuthErrorPage({ reset }: AuthErrorPageProps) {
  const t = useTranslations('error');

  return (
    <div className="w-full max-w-md rounded-lg bg-card p-8 text-center">
      <h1 className="mb-2 text-xl font-bold text-foreground">{t('title')}</h1>
      <p className="mb-6 text-sm text-muted-foreground">{t('description')}</p>
      <button
        onClick={reset}
        className="rounded-md bg-primary px-6 py-2 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        {t('tryAgain')}
      </button>
    </div>
  );
}
