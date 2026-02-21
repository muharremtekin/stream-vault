'use client';

import { useState, useCallback } from 'react';

import { cn } from '@/lib/utils/cn';
import { Badge } from '@/components/ui/badge';

interface TagInputProps {
  label?: string;
  value: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
  error?: string;
  className?: string;
}

export function TagInput({
  label,
  value,
  onChange,
  placeholder,
  error,
  className,
}: TagInputProps) {
  const [input, setInput] = useState('');

  const addTag = useCallback(
    (raw: string) => {
      const tag = raw.trim();
      if (!tag) return;
      if (value.includes(tag)) return;
      onChange([...value, tag]);
      setInput('');
    },
    [value, onChange]
  );

  const removeTag = useCallback(
    (index: number) => {
      onChange(value.filter((_, i) => i !== index));
    },
    [value, onChange]
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      addTag(input);
    }
    if (e.key === 'Backspace' && !input && value.length > 0) {
      removeTag(value.length - 1);
    }
  };

  return (
    <div className={cn('w-full', className)}>
      {label && (
        <label className="mb-1.5 block text-sm font-medium text-foreground">
          {label}
        </label>
      )}
      <div
        className={cn(
          'flex flex-wrap items-center gap-1.5 rounded-md border bg-input px-3 py-2 transition-colors',
          'focus-within:ring-2 focus-within:ring-ring focus-within:border-ring',
          error ? 'border-destructive' : 'border-input-border'
        )}
      >
        {value.map((tag, i) => (
          <Badge key={tag} variant="default" size="sm" isRemovable onRemove={() => removeTag(i)}>
            {tag}
          </Badge>
        ))}
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={() => addTag(input)}
          placeholder={value.length === 0 ? placeholder : undefined}
          className="min-w-24 flex-1 border-none bg-transparent text-sm text-foreground placeholder:text-muted-foreground outline-none"
        />
      </div>
      {error && <p className="mt-1.5 text-xs text-destructive">{error}</p>}
    </div>
  );
}
