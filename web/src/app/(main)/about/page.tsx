import type { Metadata } from 'next';

import Link from 'next/link';
import { Wifi, Users, Globe, Sparkles } from 'lucide-react';
import { getTranslations } from 'next-intl/server';

import { FeatureCard } from '@/components/landing/feature-card';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('pages.about');
  return {
    title: t('title'),
    description: t('description'),
  };
}

export default async function AboutPage() {
  const t = await getTranslations('pages.about');

  const features = [
    {
      icon: <Wifi className="h-6 w-6" />,
      titleKey: 'featureAdaptiveTitle' as const,
      descKey: 'featureAdaptiveDesc' as const,
    },
    {
      icon: <Users className="h-6 w-6" />,
      titleKey: 'featureMultiProfileTitle' as const,
      descKey: 'featureMultiProfileDesc' as const,
    },
    {
      icon: <Globe className="h-6 w-6" />,
      titleKey: 'featureMultiLangTitle' as const,
      descKey: 'featureMultiLangDesc' as const,
    },
    {
      icon: <Sparkles className="h-6 w-6" />,
      titleKey: 'featureCuratedTitle' as const,
      descKey: 'featureCuratedDesc' as const,
    },
  ];

  const stats = [
    { valueKey: 'statContentLibraryValue' as const, labelKey: 'statContentLibrary' as const },
    { valueKey: 'statDevicesValue' as const, labelKey: 'statDevices' as const },
    { valueKey: 'statQualityValue' as const, labelKey: 'statQuality' as const },
    { valueKey: 'statLanguagesValue' as const, labelKey: 'statLanguages' as const },
  ];

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:py-12">
      {/* Hero Section */}
      <section className="relative mb-16 overflow-hidden rounded-xl border border-border bg-card px-6 py-16 text-center sm:px-12 sm:py-20">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(229,9,20,0.15),transparent_50%)]" />
        <div className="relative">
          <h1 className="mb-4 text-3xl font-bold text-foreground sm:text-4xl lg:text-5xl">
            {t('title')}
          </h1>
          <p className="mx-auto max-w-2xl text-base text-muted-foreground sm:text-lg">
            {t('description')}
          </p>
        </div>
      </section>

      {/* Mission Section */}
      <section className="mb-16 text-center">
        <h2 className="mb-4 text-2xl font-bold text-foreground">{t('missionTitle')}</h2>
        <p className="mx-auto max-w-3xl text-muted-foreground">{t('missionDescription')}</p>
      </section>

      {/* Features Grid */}
      <section className="mb-16">
        <h2 className="mb-8 text-center text-2xl font-bold text-foreground">
          {t('featuresTitle')}
        </h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          {features.map((feature) => (
            <FeatureCard
              key={feature.titleKey}
              icon={feature.icon}
              title={t(feature.titleKey)}
              description={t(feature.descKey)}
            />
          ))}
        </div>
      </section>

      {/* Stats Section */}
      <section className="mb-16">
        <h2 className="mb-8 text-center text-2xl font-bold text-foreground">
          {t('statsTitle')}
        </h2>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {stats.map((stat) => (
            <div
              key={stat.labelKey}
              className="rounded-lg border border-border bg-card p-6 text-center"
            >
              <p className="text-2xl font-bold text-primary sm:text-3xl">
                {t(stat.valueKey)}
              </p>
              <p className="mt-1 text-sm text-muted-foreground">{t(stat.labelKey)}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Team Section */}
      <section className="text-center">
        <h2 className="mb-4 text-2xl font-bold text-foreground">{t('teamTitle')}</h2>
        <p className="mx-auto mb-6 max-w-3xl text-muted-foreground">
          {t('teamDescription')}
        </p>
        <Link
          href="/help"
          className="inline-block rounded-md bg-primary px-6 py-2.5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        >
          {t('contactCta')}
        </Link>
      </section>
    </div>
  );
}
