'use client';

import { useRouter } from 'next/navigation';

import { useTranslations } from 'next-intl';

import { Button } from '@/components/ui/button';

interface SearchErrorProps {
  error: Error;
  reset: () => void;
}

export default function SearchError({ error: _error, reset }: SearchErrorProps) {
  const router = useRouter();
  const t = useTranslations('error');

  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center gap-6 px-4 text-center">
      <h1 className="text-2xl font-bold text-foreground">{t('title')}</h1>
      <p className="text-muted-foreground">{t('description')}</p>
      <div className="flex gap-3">
        <Button onClick={reset}>{t('tryAgain')}</Button>
        <Button variant="ghost" onClick={() => router.push('/browse')}>
          {t('goHome')}
        </Button>
      </div>
    </div>
  );
}
