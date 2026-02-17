import { cookies } from 'next/headers';
import Link from 'next/link';
import { redirect } from 'next/navigation';
import { getTranslations } from 'next-intl/server';

import { HeroSection } from '@/components/landing/hero-section';
import { FeatureGrid } from '@/components/landing/feature-grid';
import { FaqAccordion } from '@/components/landing/faq-accordion';

export default async function LandingPage() {
  const cookieStore = await cookies();
  const token = cookieStore.get('sv-access-token');
  if (token?.value) {
    redirect('/browse');
  }

  const t = await getTranslations('landing');

  return (
    <main className="flex min-h-screen flex-col bg-background">
      <HeroSection />
      <FeatureGrid />
      <FaqAccordion />
      <section className="mx-auto max-w-2xl px-4 py-16 text-center">
        <h2 className="mb-2 text-2xl font-bold text-foreground sm:text-3xl">
          {t('readyToWatch')}
        </h2>
        <p className="mb-6 text-muted-foreground">{t('readyToWatchDesc')}</p>
        <Link
          href="/register"
          className="inline-block rounded-md bg-primary px-8 py-3 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none sm:text-base"
        >
          {t('getStarted')}
        </Link>
      </section>
    </main>
  );
}
