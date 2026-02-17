'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

import { z } from 'zod';
import { Mail } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const heroEmailSchema = z.object({
  email: z.string().min(1).email(),
});

export function HeroEmailForm() {
  const t = useTranslations('landing');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');

    const result = heroEmailSchema.safeParse({ email });
    if (!result.success) {
      setError(result.error.issues[0]?.message ?? 'Invalid email');
      return;
    }

    setIsLoading(true);
    router.push(`/register?email=${encodeURIComponent(email)}`);
  };

  return (
    <form onSubmit={handleSubmit} noValidate className="mt-8 w-full max-w-xl">
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={t('emailPlaceholder')}
            autoComplete="email"
            className={cn(
              'h-12 w-full rounded-md border bg-card pl-10 pr-4 text-sm text-foreground',
              'placeholder:text-muted-foreground',
              'focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
              'transition-colors',
              error ? 'border-destructive' : 'border-border'
            )}
          />
        </div>
        <button
          type="submit"
          disabled={isLoading}
          className={cn(
            'h-12 rounded-md bg-primary px-8 text-sm font-semibold text-primary-foreground',
            'transition-colors hover:bg-accent',
            'focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
            'disabled:cursor-not-allowed disabled:opacity-60',
            'whitespace-nowrap'
          )}
        >
          {t('getStarted')}
        </button>
      </div>
      {error && (
        <p className="mt-1.5 text-xs text-destructive">{error}</p>
      )}
    </form>
  );
}
