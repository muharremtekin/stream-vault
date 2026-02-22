import type { Metadata } from 'next';

import { getTranslations } from 'next-intl/server';
import { Mail } from 'lucide-react';

import { HelpFaqAccordion } from '@/components/help/help-faq-accordion';

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations('pages.help');
  return {
    title: t('title'),
  };
}

export default async function HelpPage() {
  const t = await getTranslations('pages.help');

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <h1 className="mb-2 text-2xl font-bold text-foreground sm:text-3xl">
        {t('title')}
      </h1>
      <p className="mb-10 text-muted-foreground">{t('description')}</p>

      <HelpFaqAccordion />

      <section className="mt-12 rounded-lg border border-border bg-card p-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10">
            <Mail className="h-5 w-5 text-primary" aria-hidden="true" />
          </div>
          <p className="text-sm text-muted-foreground sm:text-base">
            {t('contact')}
          </p>
        </div>
      </section>
    </div>
  );
}
