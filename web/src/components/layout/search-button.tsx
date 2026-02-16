'use client';

import { useRouter } from 'next/navigation';

import { Search } from 'lucide-react';
import { useTranslations } from 'next-intl';

export function SearchButton() {
  const router = useRouter();
  const t = useTranslations('layout.navbar');

  return (
    <button
      onClick={() => router.push('/search')}
      aria-label={t('search')}
      className="rounded-full p-2 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <Search className="h-5 w-5" />
    </button>
  );
}
