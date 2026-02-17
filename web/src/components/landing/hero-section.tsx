import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

import { HeroEmailForm } from '@/components/landing/hero-email-form';

export async function HeroSection() {
  const tLanding = await getTranslations('landing');
  const tAuth = await getTranslations('auth');

  return (
    <section className="relative flex min-h-screen flex-col items-center justify-center px-4 text-center">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(229,9,20,0.15),transparent_50%)]" />
      <div className="relative z-10 flex flex-col items-center">
        <h1 className="mb-4 text-5xl font-bold tracking-tight text-primary sm:text-6xl lg:text-7xl">
          StreamVault
        </h1>
        <p className="mx-auto max-w-2xl text-lg text-foreground sm:text-xl lg:text-2xl">
          {tLanding('heroTitle')}
        </p>
        <p className="mt-2 text-sm text-muted-foreground sm:text-base">
          {tLanding('heroDescription')}
        </p>
        <HeroEmailForm />
        <p className="mt-4 text-sm text-muted-foreground">
          {tAuth('hasAccount')}{' '}
          <Link
            href="/login"
            className="font-medium text-primary underline-offset-4 hover:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            {tLanding('signIn')}
          </Link>
        </p>
      </div>
    </section>
  );
}
