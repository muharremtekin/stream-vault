'use client';

import { useId } from 'react';

import { cn } from '@/lib/utils/cn';

const INPUT_SIZES = {
  sm: 'h-8 text-xs',
  md: 'h-10 text-sm',
  lg: 'h-12 text-base',
} as const;

type InputSize = keyof typeof INPUT_SIZES;

interface InputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  error?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  size?: InputSize;
  ref?: React.Ref<HTMLInputElement>;
}

export function Input({
  label,
  error,
  helperText,
  leftIcon,
  rightIcon,
  size = 'md',
  ref,
  className,
  id: providedId,
  name,
  disabled,
  ...rest
}: InputProps) {
  const generatedId = useId();
  const inputId = providedId ?? (name ? `input-${name}` : generatedId);
  const errorId = error ? `${inputId}-error` : undefined;
  const helperTextId = helperText && !error ? `${inputId}-helper` : undefined;

  return (
    <div className={cn('w-full', className)}>
      {label && (
        <label
          htmlFor={inputId}
          className="mb-1.5 block text-sm font-medium text-foreground"
        >
          {label}
        </label>
      )}

      <div className="relative">
        {leftIcon && (
          <div className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
            {leftIcon}
          </div>
        )}

        <input
          ref={ref}
          id={inputId}
          name={name}
          disabled={disabled}
          aria-invalid={error ? true : undefined}
          aria-describedby={errorId ?? helperTextId}
          className={cn(
            'w-full rounded-md border bg-input px-3 text-foreground placeholder:text-muted-foreground transition-colors',
            'focus-visible:ring-2 focus-visible:ring-ring focus-visible:border-ring focus-visible:outline-none',
            INPUT_SIZES[size],
            error
              ? 'border-destructive focus-visible:ring-destructive'
              : 'border-input-border',
            leftIcon && 'pl-10',
            rightIcon && 'pr-10',
            disabled && 'cursor-not-allowed opacity-50'
          )}
          {...rest}
        />

        {rightIcon && (
          <div className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">
            {rightIcon}
          </div>
        )}
      </div>

      {error && (
        <p id={errorId} className="mt-1.5 text-xs text-destructive">
          {error}
        </p>
      )}

      {helperText && !error && (
        <p id={helperTextId} className="mt-1.5 text-xs text-muted-foreground">
          {helperText}
        </p>
      )}
    </div>
  );
}
