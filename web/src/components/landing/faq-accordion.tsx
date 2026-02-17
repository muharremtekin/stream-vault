'use client';

import { useState } from 'react';

import { ChevronDown } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

interface FaqItemProps {
  question: string;
  answer: string;
  isOpen: boolean;
  onToggle: () => void;
}

function FaqItem({ question, answer, isOpen, onToggle }: FaqItemProps) {
  return (
    <div className="border-b border-border">
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center justify-between py-4 text-left text-foreground transition-colors hover:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
        aria-expanded={isOpen}
      >
        <span className="text-base font-medium sm:text-lg">{question}</span>
        <ChevronDown
          className={cn(
            'h-5 w-5 shrink-0 transition-transform duration-200',
            isOpen && 'rotate-180'
          )}
        />
      </button>
      <div
        className={cn(
          'overflow-hidden transition-all duration-200',
          isOpen ? 'max-h-96 pb-4' : 'max-h-0'
        )}
      >
        <p className="text-sm text-muted-foreground sm:text-base">{answer}</p>
      </div>
    </div>
  );
}

const FAQ_KEYS = ['faq1', 'faq2', 'faq3', 'faq4'] as const;

export function FaqAccordion() {
  const t = useTranslations('landing');
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  const handleToggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <section className="mx-auto max-w-3xl px-4 py-16">
      <h2 className="mb-8 text-center text-2xl font-bold text-foreground sm:text-3xl">
        {t('faqTitle')}
      </h2>
      <div className="border-t border-border">
        {FAQ_KEYS.map((key, index) => (
          <FaqItem
            key={key}
            question={t(`${key}q`)}
            answer={t(`${key}a`)}
            isOpen={openIndex === index}
            onToggle={() => handleToggle(index)}
          />
        ))}
      </div>
    </section>
  );
}
