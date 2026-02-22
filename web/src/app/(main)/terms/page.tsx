import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';

const SECTION_KEYS = [
  'acceptance',
  'account',
  'contentUsage',
  'subscription',
  'intellectualProperty',
  'termination',
  'liability',
] as const;

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('pages.terms');
  return {
    title: t('title'),
    description: t('description'),
  };
}

export default async function TermsPage() {
  const t = await getTranslations('pages.terms');

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <h1 className="mb-4 text-2xl font-bold text-foreground">{t('title')}</h1>
      <p className="mb-2 text-muted-foreground">{t('description')}</p>
      <p className="mb-8 text-sm text-muted-foreground">{t('lastUpdated')}</p>

      <div className="space-y-8">
        {SECTION_KEYS.map((key) => (
          <section key={key}>
            <h2 className="mb-3 text-xl font-semibold text-foreground">
              {t(`sections.${key}.title`)}
            </h2>
            <p className="leading-relaxed text-muted-foreground">
              {t(`sections.${key}.content`)}
            </p>
          </section>
        ))}
      </div>
    </div>
  );
}
