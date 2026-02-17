import { Monitor, Download, ShieldOff, Baby } from 'lucide-react';
import { getTranslations } from 'next-intl/server';

import { FeatureCard } from '@/components/landing/feature-card';

export async function FeatureGrid() {
  const t = await getTranslations('landing');

  const features = [
    {
      icon: <Monitor className="h-6 w-6" />,
      title: t('featureStreamTitle'),
      description: t('featureStreamDesc'),
    },
    {
      icon: <Download className="h-6 w-6" />,
      title: t('featureDownloadTitle'),
      description: t('featureDownloadDesc'),
    },
    {
      icon: <ShieldOff className="h-6 w-6" />,
      title: t('featureNoAdsTitle'),
      description: t('featureNoAdsDesc'),
    },
    {
      icon: <Baby className="h-6 w-6" />,
      title: t('featureKidsTitle'),
      description: t('featureKidsDesc'),
    },
  ];

  return (
    <section className="mx-auto max-w-4xl px-4 py-16">
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
        {features.map((feature) => (
          <FeatureCard
            key={feature.title}
            icon={feature.icon}
            title={feature.title}
            description={feature.description}
          />
        ))}
      </div>
    </section>
  );
}
