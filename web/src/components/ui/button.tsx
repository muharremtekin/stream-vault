'use client';

import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const BUTTON_VARIANTS = {
  primary: 'bg-primary text-primary-foreground hover:bg-accent',
  secondary: 'border border-border bg-card text-foreground hover:bg-muted',
  ghost: 'text-muted-foreground hover:bg-muted hover:text-foreground',
  danger: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
  link: 'text-primary underline-offset-4 hover:underline',
} as const;

type ButtonVariant = keyof typeof BUTTON_VARIANTS;

const BUTTON_SIZES = {
  sm: 'h-8 px-3 text-xs',
  md: 'h-10 px-4 text-sm',
  lg: 'h-12 px-6 text-base',
} as const;

type ButtonSize = keyof typeof BUTTON_SIZES;

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  isLoading?: boolean;
  href?: string;
  ref?: React.Ref<HTMLButtonElement>;
}

function Spinner() {
  return (
    <svg
      className="h-4 w-4 animate-spin"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  );
}

const BASE_CLASSES =
  'inline-flex items-center justify-center gap-2 rounded-md font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50 disabled:pointer-events-none';

export function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  href,
  ref,
  className,
  disabled,
  children,
  ...rest
}: ButtonProps) {
  const t = useTranslations('ui.button');

  const classes = cn(
    BASE_CLASSES,
    BUTTON_VARIANTS[variant],
    BUTTON_SIZES[size],
    className
  );

  if (href) {
    return (
      <a href={href} className={classes}>
        {children}
      </a>
    );
  }

  return (
    <button
      ref={ref}
      className={classes}
      disabled={disabled || isLoading}
      aria-busy={isLoading || undefined}
      aria-disabled={disabled || isLoading || undefined}
      {...rest}
    >
      {isLoading && (
        <>
          <Spinner />
          <span className="sr-only">{t('loading')}</span>
        </>
      )}
      {children}
    </button>
  );
}
