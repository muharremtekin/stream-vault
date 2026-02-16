import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

export default async function LandingPage() {
  const t = await getTranslations('landing');

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 px-4">
      <div className="text-center">
        <h1 className="mb-4 text-5xl font-bold tracking-tight text-primary sm:text-6xl">
          StreamVault
        </h1>
        <p className="text-lg text-muted-foreground sm:text-xl">
          {t('heroTitle')}
        </p>
        <p className="mt-2 text-sm text-muted-foreground">
          {t('heroDescription')}
        </p>
      </div>

      <div className="flex gap-4">
        <Link
          href="/login"
          className="rounded-md bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t('signIn')}
        </Link>
        <Link
          href="/browse"
          className="rounded-md border border-border bg-card px-6 py-3 text-sm font-semibold text-foreground transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t('getStarted')}
        </Link>
      </div>
    </div>
  );
}
