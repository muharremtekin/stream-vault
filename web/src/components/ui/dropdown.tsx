'use client';

import { useState, useRef, useEffect, useCallback } from 'react';

import { cn } from '@/lib/utils/cn';

const DROPDOWN_ALIGNS = {
  left: 'left-0',
  right: 'right-0',
  center: 'left-1/2 -translate-x-1/2',
} as const;

type DropdownAlign = keyof typeof DROPDOWN_ALIGNS;

interface DropdownProps {
  trigger: React.ReactNode;
  children: React.ReactNode;
  align?: DropdownAlign;
}

export function Dropdown({ trigger, children, align = 'left' }: DropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node) // DOM event cast
      ) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      if (!isOpen) return;

      if (event.key === 'Escape') {
        setIsOpen(false);
        return;
      }

      if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
        event.preventDefault();
        const items = menuRef.current?.querySelectorAll<HTMLElement>('[role="menuitem"]:not([aria-disabled="true"])');
        if (!items?.length) return;

        const current = document.activeElement;
        const currentIndex = Array.from(items).indexOf(current as HTMLElement);
        const nextIndex =
          event.key === 'ArrowDown'
            ? (currentIndex + 1) % items.length
            : (currentIndex - 1 + items.length) % items.length;
        items[nextIndex].focus();
      }
    },
    [isOpen]
  );

  return (
    <div ref={containerRef} className="relative inline-block" onKeyDown={handleKeyDown}>
      <div
        role="button"
        tabIndex={0}
        aria-haspopup="menu"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((prev) => !prev)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            setIsOpen((prev) => !prev);
          }
        }}
      >
        {trigger}
      </div>

      {isOpen && (
        <div
          ref={menuRef}
          role="menu"
          className={cn(
            'absolute z-50 mt-2 min-w-48 rounded-md border border-border bg-card py-1 shadow-lg',
            DROPDOWN_ALIGNS[align]
          )}
        >
          {children}
        </div>
      )}
    </div>
  );
}

interface DropdownItemProps {
  onClick?: () => void;
  icon?: React.ReactNode;
  children: React.ReactNode;
  isDisabled?: boolean;
}

export function DropdownItem({
  onClick,
  icon,
  children,
  isDisabled = false,
}: DropdownItemProps) {
  return (
    <button
      type="button"
      role="menuitem"
      tabIndex={-1}
      disabled={isDisabled}
      aria-disabled={isDisabled || undefined}
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-muted-foreground transition-colors',
        'hover:bg-muted hover:text-foreground',
        'focus-visible:bg-muted focus-visible:text-foreground focus-visible:outline-none',
        isDisabled && 'cursor-not-allowed opacity-50'
      )}
    >
      {icon && <span className="h-4 w-4 shrink-0">{icon}</span>}
      {children}
    </button>
  );
}

interface DropdownDividerProps {
  className?: string;
}

export function DropdownDivider({ className }: DropdownDividerProps) {
  return <div className={cn('my-1 border-t border-border', className)} />;
}
