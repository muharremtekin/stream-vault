'use client';

import { useState } from 'react';

import { ChevronDown } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const FAQ_CATEGORIES = ['account', 'streaming', 'subscription', 'technical'] as const;

type FaqCategory = (typeof FAQ_CATEGORIES)[number];

const QUESTION_KEYS = ['q1', 'q2', 'q3'] as const;

interface FaqItemProps {
  question: string;
  answer: string;
  isOpen: boolean;
  onToggle: () => void;
}

function FaqItem({ question, answer, isOpen, onToggle }: FaqItemProps) {
  return (
    <div className="border-b border-border last:border-b-0">
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center justify-between py-4 text-left text-foreground transition-colors hover:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        aria-expanded={isOpen}
      >
        <span className="text-sm font-medium sm:text-base">{question}</span>
        <ChevronDown
          className={cn(
            'ml-2 h-5 w-5 shrink-0 transition-transform duration-200',
            isOpen && 'rotate-180'
          )}
          aria-hidden="true"
        />
      </button>
      <div
        className={cn(
          'overflow-hidden transition-all duration-200',
          isOpen ? 'max-h-96 pb-4' : 'max-h-0'
        )}
      >
        <p className="text-sm text-muted-foreground">{answer}</p>
      </div>
    </div>
  );
}

export function HelpFaqAccordion() {
  const t = useTranslations('pages.help.faq');
  const [openKey, setOpenKey] = useState<string | null>(null);

  const handleToggle = (key: string) => {
    setOpenKey(openKey === key ? null : key);
  };

  return (
    <section aria-label={t('title')}>
      <h2 className="mb-6 text-xl font-bold text-foreground sm:text-2xl">
        {t('title')}
      </h2>
      <div className="space-y-8">
        {FAQ_CATEGORIES.map((category: FaqCategory) => (
          <div key={category}>
            <h3 className="mb-3 text-lg font-semibold text-foreground">
              {t(`categories.${category}`)}
            </h3>
            <div className="rounded-lg border border-border bg-card">
              <div className="px-4">
                {QUESTION_KEYS.map((qKey) => {
                  const itemKey = `${category}-${qKey}`;
                  const aKey = qKey.replace('q', 'a');
                  return (
                    <FaqItem
                      key={itemKey}
                      question={t(`${category}.${qKey}`)}
                      answer={t(`${category}.${aKey}`)}
                      isOpen={openKey === itemKey}
                      onToggle={() => handleToggle(itemKey)}
                    />
                  );
                })}
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
