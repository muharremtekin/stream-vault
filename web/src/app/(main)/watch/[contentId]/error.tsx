'use client';

import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { Button } from '@/components/ui/button';

interface WatchErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function WatchError({ reset }: WatchErrorProps) {
  const t = useTranslations('error');
  const tPlayer = useTranslations('player');
  const router = useRouter();

  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center gap-6 bg-black text-white">
      <p className="text-lg font-medium">{t('title')}</p>
      <p className="text-sm text-white/60">{t('description')}</p>
      <div className="flex gap-3">
        <Button variant="secondary" onClick={() => router.back()}>
          {tPlayer('goBack')}
        </Button>
        <Button onClick={reset}>{t('tryAgain')}</Button>
      </div>
    </div>
  );
}
