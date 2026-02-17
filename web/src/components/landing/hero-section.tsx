import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

export async function HeroSection() {
  const t = await getTranslations('landing');

  return (
    <section className="relative flex min-h-screen flex-col items-center justify-center px-4 text-center">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(229,9,20,0.15),transparent_50%)]" />
      <div className="relative z-10">
        <h1 className="mb-4 text-5xl font-bold tracking-tight text-primary sm:text-6xl lg:text-7xl">
          StreamVault
        </h1>
        <p className="mx-auto max-w-2xl text-lg text-foreground sm:text-xl lg:text-2xl">
          {t('heroTitle')}
        </p>
        <p className="mt-2 text-sm text-muted-foreground sm:text-base">
          {t('heroDescription')}
        </p>
        <div className="mt-8 flex justify-center gap-4">
          <Link
            href="/login"
            className="rounded-md bg-primary px-8 py-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none sm:text-base"
          >
            {t('signIn')}
          </Link>
          <Link
            href="/register"
            className="rounded-md border border-border bg-card px-8 py-3 text-sm font-semibold text-foreground transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none sm:text-base"
          >
            {t('getStarted')}
          </Link>
        </div>
      </div>
    </section>
  );
}
